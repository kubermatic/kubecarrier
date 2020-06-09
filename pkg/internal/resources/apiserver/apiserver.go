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
	"strings"

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
	resources := []string{"../default"}
	if c.Spec.TLSSecretRef == nil || c.Spec.TLSSecretRef.Name == "" {
		// TLS Secret is not present, we create a self-signed certificate for localhost via cert-manager.
		resources = append(resources, "../certmanager")
		c.Spec.TLSSecretRef = &operatorv1alpha1.ObjectReference{
			Name: "tls-server-cert",
		}
	}
	// Patch environment
	// Note:
	// we are not using *appsv1.Deployment here,
	// because some fields will be defaulted to empty and
	// interfere with the strategic merge patch of kustomize.
	var managerEnv map[string]interface{}
	if c.Spec.OIDC != nil {
		extraArgs := make([]string, 0)
		if len(c.Spec.OIDC.RequiredClaims) > 0 {
			rclaims := make([]string, 0)
			for k, v := range c.Spec.OIDC.RequiredClaims {
				rclaims = append(rclaims, fmt.Sprintf("%s=%s", k, v))
			}
			extraArgs = append(extraArgs, "--oidc-required-claim="+strings.Join(rclaims, ","))
		}

		managerEnv = map[string]interface{}{
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
								"args": append([]string{
									"--address=$(API_SERVER_ADDR)",
									"--tls-cert-file=$(API_SERVER_TLS_CERT_FILE)",
									"--tls-private-key-file=$(API_SERVER_TLS_PRIVATE_KEY_FILE)",
									"--oidc-issuer-url=$(API_SERVER_OIDC_ISSUER_URL)",
									"--oidc-client-id=$(API_SERVER_OIDC_CLIENT_ID)",
									"--oidc-ca-file=$(API_SERVER_OIDC_CA_FILE)",
									"--oidc-username-claim=$(API_SERVER_OIDC_USERNAME_CLAIM)",
									"--oidc-username-prefix=$(API_SERVER_OIDC_USERNAME_PREFIX)",
									"--oidc-groups-claim=$(API_SERVER_OIDC_GROUPS_CLAIM)",
									"--oidc-groups-prefix=$(API_SERVER_OIDC_GROUPS_PREFIX)",
									"--oidc-signing-algs=$(API_SERVER_OIDC_SIGNING_ALGS)",
								}, extraArgs...),
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
									{
										"name":  "API_SERVER_OIDC_ISSUER_URL",
										"value": c.Spec.OIDC.IssuerURL,
									},
									{
										"name":  "API_SERVER_OIDC_CLIENT_ID",
										"value": c.Spec.OIDC.ClientID,
									},
									{
										"name":  "API_SERVER_OIDC_CA_FILE",
										"value": "/run/oidc-certs/ca.crt",
									},
									{
										"name":  "API_SERVER_OIDC_USERNAME_CLAIM",
										"value": c.Spec.OIDC.UsernameClaim,
									},
									{
										"name":  "API_SERVER_OIDC_USERNAME_PREFIX",
										"value": c.Spec.OIDC.UsernamePrefix,
									},
									{
										"name":  "API_SERVER_OIDC_GROUPS_CLAIM",
										"value": c.Spec.OIDC.GroupsClaim,
									},
									{
										"name":  "API_SERVER_OIDC_GROUPS_PREFIX",
										"value": c.Spec.OIDC.GroupsPrefix,
									},
									{
										"name":  "API_SERVER_OIDC_SIGNING_ALGS",
										"value": strings.Join(c.Spec.OIDC.SupportedSigningAlgs, ","),
									},
								},
								"volumeMounts": []map[string]interface{}{
									{
										"mountPath": "/run/serving-certs",
										"readyOnly": true,
										"name":      "serving-cert",
									},
									{
										"mountPath": "/run/oidc-certs",
										"readyOnly": true,
										"name":      "oidc-cert",
									},
								},
							},
						},
						"volumes": []map[string]interface{}{
							{
								"name": "serving-cert",
								"secret": map[string]interface{}{
									"secretName": c.Spec.TLSSecretRef.Name,
								},
							},
							{
								"name": "oidc-cert",
								"secret": map[string]interface{}{
									"secretName": c.Spec.OIDC.CertificateAuthority.Name,
								},
							},
						},
					},
				},
			},
		}
	} else {
		managerEnv = map[string]interface{}{
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
									"--authentication-mode=$(AUTHENTICATION_MODE)",
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
									{
										"name":  "AUTHENTICATION_MODE",
										"value": "Anonymous",
									},
								},
								"volumeMounts": []map[string]interface{}{
									{
										"mountPath": "/run/serving-certs",
										"readyOnly": true,
										"name":      "serving-cert",
									},
								},
							},
						},
						"volumes": []map[string]interface{}{
							{
								"name": "serving-cert",
								"secret": map[string]interface{}{
									"secretName": c.Spec.TLSSecretRef.Name,
								},
							},
						},
					},
				},
			},
		}

	}
	managerEnvBytes, err := yaml.Marshal(managerEnv)
	if err != nil {
		return nil, fmt.Errorf("marshalling manager env patch: %w", err)
	}
	if err = kc.WriteFile("/man/manager_env_patch.yaml", managerEnvBytes); err != nil {
		return nil, fmt.Errorf("writing manager_env_patch.yaml: %w", err)
	}

	if err := kc.MkLayer("man", types.Kustomization{
		Namespace: c.Namespace,
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/apiserver",
				NewTag: v.Version,
			},
		},
		PatchesStrategicMerge: []types.PatchStrategicMerge{
			"manager_env_patch.yaml",
		},
		Resources: resources,
	}); err != nil {
		return nil, fmt.Errorf("cannot mkdir: %w", err)
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
		labels[constants.NameLabel] = "apiserver"
		labels[constants.InstanceLabel] = c.Name
		labels[constants.ManagedbyLabel] = constants.ManagedbyKubeCarrierOperator
		labels[constants.VersionLabel] = v.Version
		obj.SetLabels(labels)
	}
	return objects, nil
}
