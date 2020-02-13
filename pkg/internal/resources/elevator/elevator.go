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

package elevator

import (
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/image"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier controller manager setup.
type Config struct {
	// Name is the name of the Elevator instance.
	Name string
	// Namespace that the Elevator instance should be deployed into.
	Namespace string

	ProviderKind, ProviderVersion, ProviderGroup, ProviderPlural string
	TenantKind, TenantVersion, TenantGroup, TenantPlural         string
	DerivedCRName                                                string
}

var k = kustomize.NewDefaultKustomize()

func Manifests(c Config) ([]unstructured.Unstructured, error) {
	v := version.Get()
	kc := k.ForHTTP(vfs)
	if err := kc.MkLayer("man", types.Kustomization{
		// "." needs to be replaced, because it's forbidden for Deployment and Pod names
		NamePrefix: strings.Replace(c.Name, ".", "-", -1) + "-",
		Namespace:  c.Namespace,
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/elevator",
				NewTag: v.Version,
			},
		},
		Resources: []string{"../default"},
		PatchesStrategicMerge: []types.PatchStrategicMerge{
			"manager_env_patch.yaml",
		},
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
							"env": []map[string]interface{}{
								{
									"name":  "ELEVATOR_DERIVED_CRD_NAME",
									"value": c.DerivedCRName,
								},
								{
									"name":  "ELEVATOR_PROVIDER_KIND",
									"value": c.ProviderKind,
								},
								{
									"name":  "ELEVATOR_PROVIDER_VERSION",
									"value": c.ProviderVersion,
								},
								{
									"name":  "ELEVATOR_PROVIDER_GROUP",
									"value": c.ProviderGroup,
								},
								{
									"name":  "ELEVATOR_TENANT_KIND",
									"value": c.TenantKind,
								},
								{
									"name":  "ELEVATOR_TENANT_VERSION",
									"value": c.TenantVersion,
								},
								{
									"name":  "ELEVATOR_TENANT_GROUP",
									"value": c.TenantGroup,
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

	// Generate ClusterRole for component
	role := rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "manager",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{c.ProviderGroup},
				Resources: []string{c.ProviderPlural},
				Verbs: []string{
					"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{c.ProviderGroup},
				Resources: []string{c.ProviderPlural + "/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
			{
				APIGroups: []string{c.TenantGroup},
				Resources: []string{c.TenantPlural},
				Verbs: []string{
					"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{c.TenantGroup},
				Resources: []string{c.TenantPlural + "/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
		},
	}
	roleBytes, err := yaml.Marshal(role)
	if err != nil {
		return nil, fmt.Errorf("marshalling cluster role: %w", err)
	}
	if err := kc.WriteFile("/rbac/cluster_role.yaml", roleBytes); err != nil {
		return nil, fmt.Errorf("writing /rbac/cluster_role.yaml: %w", err)
	}

	// execute kustomize
	objects, err := kc.Build("/man")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
