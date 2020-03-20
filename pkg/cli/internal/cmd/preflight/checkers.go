/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package preflight

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/gernest/wow"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"github.com/kubermatic/kubecarrier/pkg/cli/internal/spinner"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(appsv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

const (
	firstSupportedKubernetesVersion = "v1.16.0"

	certManagerNamespace            = "cert-manager"
	certManagerCAInjectorDeployment = "cert-manager-cainjector"
	certManagerWebhookDeployment    = "cert-manager-webhook"
	certManagerDeployment           = "cert-manager"
)

// checker checks if the state of the system meets KubeCarrier installation requirements
type checker interface {
	check() error
	name() string
}

func RunCheckers(c *rest.Config, s *wow.Wow, startTime time.Time, log logr.Logger) error {
	var errBuffer bytes.Buffer
	checkers := []checker{
		&kubernetesVersionChecker{
			config:                c,
			firstSupportedVersion: firstSupportedKubernetesVersion,
		},
		&certManagerChecker{
			config:                 c,
			log:                    log,
			certManagerNamespace:   certManagerNamespace,
			certManagerDeployments: []string{certManagerDeployment, certManagerCAInjectorDeployment, certManagerWebhookDeployment},
		},
	}
	for _, checker := range checkers {
		if err := spinner.AttachSpinnerTo(s, startTime, fmt.Sprintf("[preflight check] %s", checker.name()), func() error {
			return checker.check()
		}); err != nil {
			errBuffer.WriteString(fmt.Errorf("[preflight check] %s: %w\n", checker.name(), err).Error())
		}
	}
	if errBuffer.Len() > 0 {
		return fmt.Errorf(errBuffer.String())
	}
	return nil
}

// kubernetesVersionChecker checks if the Kubernetes version of the cluster meets the requirement to deploy KubeCarrier.
type kubernetesVersionChecker struct {
	config                *rest.Config
	firstSupportedVersion string
}

func (c *kubernetesVersionChecker) check() error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c.config)
	if err != nil {
		return fmt.Errorf("cannot create discovery client: %w", err)
	}
	kubernetesVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return fmt.Errorf("can not get the kubernetesVersion: %w", err)
	}
	firstSupportedVersion, err := versionutil.ParseSemantic(c.firstSupportedVersion)
	if err != nil {
		return err
	}
	kubernetesGitVersion, err := versionutil.ParseSemantic(kubernetesVersion.String())
	if err != nil {
		return err
	}
	if kubernetesGitVersion.LessThan(firstSupportedVersion) {
		return fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= %s, found: %s", firstSupportedVersion.String(), kubernetesGitVersion.String())
	}
	return nil
}

func (c *kubernetesVersionChecker) name() string {
	return "KubernetesVersion"
}

// certManager checks if the cert-manager related deployments are ready.
type certManagerChecker struct {
	config                 *rest.Config
	log                    logr.Logger
	certManagerNamespace   string
	certManagerDeployments []string
}

func (c *certManagerChecker) check() error {
	// Get a client from the configuration of the kubernetes cluster.
	var errBuffer bytes.Buffer
	ctx := context.Background()
	client, err := util.NewClientWatcher(c.config, scheme, c.log)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	namespace := &corev1.Namespace{}
	if err := client.Get(ctx, types.NamespacedName{
		Name: c.certManagerNamespace,
	}, namespace); err != nil {
		errBuffer.WriteString(err.Error() + "\n")
	}

	for _, deploymentName := range c.certManagerDeployments {
		deployment := &appsv1.Deployment{}
		if err := client.Get(ctx, types.NamespacedName{
			Name:      deploymentName,
			Namespace: c.certManagerNamespace,
		}, deployment); err != nil {
			errBuffer.WriteString(err.Error() + "\n")
		} else {
			if !util.DeploymentIsAvailable(deployment) {
				errBuffer.WriteString(fmt.Sprintf("deployment %s is not available\n", deployment.Name))
			}
		}
	}
	if errBuffer.Len() > 0 {
		return fmt.Errorf(errBuffer.String())
	}
	return nil
}

func (c *certManagerChecker) name() string {
	return "CertManager"
}
