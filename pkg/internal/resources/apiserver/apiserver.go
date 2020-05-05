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

package apiserver

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/image"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/resources/constants"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the KubeCarrier master controller manager setup.
type Config struct {
	// Namespace is the KubeCarrier master controller manager should be deployed into.
	Namespace string
	// Name of this KubeCarrier API object.
	Name string

	// Spec of the APIServer
	Spec operatorv1alpha1.APIServerSpec
}

var k = kustomize.NewDefaultKustomize()

// Manifests generate all required manifests for the API Server
func Manifests(c Config) ([]unstructured.Unstructured, error) {
	v := version.Get()
	kc := k.ForHTTP(vfs)
	if err := kc.MkLayer("man", types.Kustomization{
		Namespace: c.Namespace,
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/api-server",
				NewTag: v.Version,
			},
		},
		PatchesStrategicMerge: []types.PatchStrategicMerge{
			"manager_env_patch.yaml",
		},
		Resources: []string{"../default"},
	}); err != nil {
		return nil, fmt.Errorf("cannot mkdir: %w", err)
	}

	// Patch environment
	// Note:
	// we are not using *appsv1.Deployment here,
	// because some fields will be defaulted to empty and
	// interfere with the strategic merge patch of kustomize.
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
							"args": []string{
								"--address=$(API_SERVER_ADDR)",
								"--tls-cert-file=$(API_SERVER_TLS_CERT_FILE)",
								"--tls-private-key-file=$(API_SERVER_TLS_PRIVATE_KEY_FILE)",
							},
							"env": []map[string]interface{}{
								{
									"name":  "API_SERVER_ADDR",
									"value": ":8443",
								},
								{
									"name":  "API_SERVER_TLS_CERT_FILE",
									"value": "/run/serving-certs/tls.crt",
								},
								{
									"name":  "API_SERVER_TLS_PRIVATE_KEY_FILE",
									"value": "/run/serving-certs/tls.key",
								},
							},
							"ports": []corev1.ContainerPort{{
								Name:          "https",
								ContainerPort: 8443,
								Protocol:      "TCP",
							}},
							"volumeMounts": []map[string]interface{}{{
								"mountPath": "/run/serving-certs",
								"readyOnly": true,
								"name":      "serving-cert",
							}},
						},
					},
					"volumes": []map[string]interface{}{{
						"name": "serving-cert",
						"secret": map[string]interface{}{
							"secretName": c.Spec.TLSSecretRef.Name,
						},
					}},
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
		labels[constants.NameLabel] = "api-server"
		labels[constants.InstanceLabel] = c.Name
		labels[constants.ManagedbyLabel] = constants.ManagedbyKubeCarrierOperator
		labels[constants.VersionLabel] = v.Version
		obj.SetLabels(labels)
	}
	return objects, nil
}
