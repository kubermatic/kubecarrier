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

package tender

import (
	"bytes"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

const (
	ServiceClusterName   = "__SERVICE_CLUSTER_NAME__"
	KubeconifgSecretName = "__KUBECONFIG_SECRET_NAME__"
)

// Config holds the config information to generate the kubecarrier tender setup.
type Config struct {
	// Namespace the tender operator should be deployed into.
	ProviderNamespace string

	// Name of this tender
	Name string

	// KubeconfigSecretName of the secret holding the service cluster kubeconfig under the "kubeconfig" key
	KubeconfigSecretName string
}

type kustomizeFactory interface {
	ForHTTP(fs http.FileSystem) kustomize.KustomizeContext
}

func Manifests(k kustomizeFactory, c Config, scheme *runtime.Scheme) ([]unstructured.Unstructured, error) {
	kc := k.ForHTTP(vfs)

	// patch settings
	kustomizePath := "/default/kustomization.yaml"
	kustomizeBytes, err := kc.ReadFile(kustomizePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", kustomizePath, err)
	}
	// kustomizeBytes = bytes.ReplaceAll(kustomizeBytes, []byte(), []byte(c.Name))
	kmap := map[string]interface{}{}
	if err := yaml.Unmarshal(kustomizeBytes, &kmap); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", kustomizePath, err)
	}

	kmap["namespace"] = c.ProviderNamespace
	kmap["namePrefix"] = fmt.Sprintf("tender-%s-", c.Name)

	v := version.Get()
	kmap["images"] = []map[string]string{
		{
			"name":   "quay.io/kubecarrier/tender",
			"newTag": v.Version,
		},
	}

	kustomizeBytes, err = yaml.Marshal(kmap)
	if err != nil {
		return nil, fmt.Errorf("remarshal %s: %w", kustomizePath, err)
	}
	if err := kc.WriteFile(kustomizePath, kustomizeBytes); err != nil {
		return nil, fmt.Errorf("writing %s: %w", kustomizePath, err)
	}

	managerPath := "/manager/manager.yaml"
	managerBytes, err := kc.ReadFile(managerPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", managerPath, err)
	}

	managerBytes = bytes.ReplaceAll(managerBytes, []byte(ServiceClusterName), []byte(c.Name))
	managerBytes = bytes.ReplaceAll(managerBytes, []byte(KubeconifgSecretName), []byte(c.KubeconfigSecretName))
	if err := kc.WriteFile(managerPath, managerBytes); err != nil {
		return nil, fmt.Errorf("writing %s: %w", managerPath, err)
	}
	// execute kustomize
	unstructuredObjects, err := kc.Build("/default")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return unstructuredObjects, nil
}
