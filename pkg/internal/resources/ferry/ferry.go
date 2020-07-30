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

package ferry

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/image"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"

	"k8c.io/kubecarrier/pkg/internal/kustomize"
	"k8c.io/kubecarrier/pkg/internal/resources/constants"
	"k8c.io/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier ferry setup.
type Config struct {
	// Namespace the ferry operator should be deployed into.
	ProviderNamespace string

	// Name of this ferry
	Name string

	// KubeconfigSecretName of the secret holding the service cluster kubeconfig under the "kubeconfig" key
	KubeconfigSecretName string
	// LogLevel
	LogLevel *int
}

var k = kustomize.NewDefaultKustomize()

// Manifests generate all required manifests for the Operator
// See https://github.com/kubermatic/kubecarrier/issues/95 for discussion
func Manifests(c Config) ([]unstructured.Unstructured, error) {
	v := version.Get()
	kc := k.ForHTTP(vfs)
	if err := kc.MkLayer("man", types.Kustomization{
		Namespace:  c.ProviderNamespace,
		NamePrefix: fmt.Sprintf("%s-", c.Name),
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/ferry",
				NewTag: v.Version,
			},
		},
		PatchesStrategicMerge: []types.PatchStrategicMerge{
			"manager_env_patch.yaml",
		},
		PatchesJson6902: []types.PatchJson6902{{
			Target: &types.PatchTarget{
				Gvk: gvk.Gvk{
					Group:   "apps",
					Kind:    "Deployment",
					Version: "v1",
				},
				Name:      "manager",
				Namespace: "system",
			},
			Patch: strings.TrimSpace(fmt.Sprintf(`
			 [
{ "op": "add", "path": "/spec/template/spec/containers/0/env/0", "value": {"name": "SERVICE_CLUSTER", "value": "%s"}},
{ "op": "add", "path": "/spec/template/spec/volumes/0/secret/secretName", "value": "%s"}
]
`, c.Name, c.KubeconfigSecretName)),
		}},
		Resources: []string{"../default"},
	}); err != nil {
		return nil, fmt.Errorf("cannot MkLayer: %w", err)
	}

	var logLevel int
	if c.LogLevel != nil {
		logLevel = *c.LogLevel
	}

	managerEnv := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]string{
			"name":      "manager",
			"namespace": "system",
		},
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name": "manager",
							"env": []map[string]interface{}{
								{
									"name":  "LOG_LEVEL",
									"value": strconv.FormatInt(int64(logLevel), 10),
								},
							},
						},
					},
				},
			},
		},
	}
	managerEnvBytes, err := yaml.Marshal(managerEnv)
	if err != nil {
		return nil, fmt.Errorf("marshalling manager env patch: %w", err)
	}
	if err = kc.WriteFile("/man/manager_env_patch.yaml", managerEnvBytes); err != nil {
		return nil, fmt.Errorf("writing manager_env_patch.yaml: %w", err)
	}

	// execute kustomize
	objects, err := kc.Build("/man")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	for _, obj := range objects {
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[constants.NameLabel] = "ferry"
		labels[constants.InstanceLabel] = c.Name
		labels[constants.ManagedByLabel] = constants.ManagedByKubeCarrierOperator
		labels[constants.VersionLabel] = v.Version
		obj.SetLabels(labels)
	}
	return objects, nil
}
