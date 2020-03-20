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
	"sigs.k8s.io/controller-runtime/pkg/client"

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

	certManagerNamespaceName            = "cert-manager"
	certManagerCAInjectorDeploymentName = "cert-manager-cainjector"
	certManagerWebhookDeploymentName    = "cert-manager-webhook"
	certManagerDeploymentName           = "cert-manager"
)

// checker checks if the state of the system meets KubeCarrier installation requirements
type checker interface {
	check() error
	name() string
}

func RunChecks(c *rest.Config, s *wow.Wow, startTime time.Time, log logr.Logger) error {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return fmt.Errorf("cannot create discovery client: %w", err)
	}
	kubernetesVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return fmt.Errorf("can not get the kubernetesVersion: %w", err)
	}
	cl, err := util.NewClientWatcher(c, scheme, log)
	if err != nil {
		return fmt.Errorf("creating Kubernetes client: %w", err)
	}
	var errBuffer bytes.Buffer
	checkers := []checker{
		&kubernetesVersionChecker{
			firstSupportedVersion: firstSupportedKubernetesVersion,
			kubernetesVersion:     kubernetesVersion.String(),
		},
		&certManagerChecker{
			client:                 cl,
			certManagerNamespace:   certManagerNamespaceName,
			certManagerDeployments: []string{certManagerDeploymentName, certManagerCAInjectorDeploymentName, certManagerWebhookDeploymentName},
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
	firstSupportedVersion string
	kubernetesVersion     string
}

func (c *kubernetesVersionChecker) check() error {
	firstSupportedVersion, err := versionutil.ParseSemantic(c.firstSupportedVersion)
	if err != nil {
		return err
	}
	kubernetesGitVersion, err := versionutil.ParseSemantic(c.kubernetesVersion)
	if err != nil {
		return err
	}
	if kubernetesGitVersion.LessThan(firstSupportedVersion) {
		return fmt.Errorf("kubernetes version is lower than the oldest version that KubeCarrier supports, requrires: >= v%s, found: v%s", firstSupportedVersion.String(), kubernetesGitVersion.String())
	}
	return nil
}

func (c *kubernetesVersionChecker) name() string {
	return "KubernetesVersion"
}

// certManager checks if the cert-manager related deployments are ready.
type certManagerChecker struct {
	client                 client.Client
	certManagerNamespace   string
	certManagerDeployments []string
}

func (c *certManagerChecker) check() error {
	// Get a client from the configuration of the kubernetes cluster.
	ctx := context.Background()
	var errBuffer bytes.Buffer
	namespace := &corev1.Namespace{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name: c.certManagerNamespace,
	}, namespace); err != nil {
		errBuffer.WriteString(err.Error() + "\n")
	}

	for _, deploymentName := range c.certManagerDeployments {
		deployment := &appsv1.Deployment{}
		if err := c.client.Get(ctx, types.NamespacedName{
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
