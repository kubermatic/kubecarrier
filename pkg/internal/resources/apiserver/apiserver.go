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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			Name: "apiserver-tls-cert",
		}
	}
	const AuthModeEnv = "AUTHENTICATION_MODE"

	deploymentPatch := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "manager",
			Namespace: "system",
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "manager",
							Args: []string{
								"--address=$(API_SERVER_ADDR)",
								"--tls-cert-file=$(API_SERVER_TLS_CERT_FILE)",
								"--tls-private-key-file=$(API_SERVER_TLS_PRIVATE_KEY_FILE)",
								"--authentication-mode=$(AUTHENTICATION_MODE)",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "API_SERVER_ADDR",
									Value: ":8443",
								},
								{
									Name:  "API_SERVER_TLS_CERT_FILE",
									Value: "/run/serving-certs/tls.crt",
								},
								{
									Name:  "API_SERVER_TLS_PRIVATE_KEY_FILE",
									Value: "/run/serving-certs/tls.key",
								},
								{
									Name:  AuthModeEnv,
									Value: "",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/run/serving-certs",
									ReadOnly:  true,
									Name:      "serving-cert",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "serving-cert",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: c.Spec.TLSSecretRef.Name,
								},
							},
						},
					},
				},
			},
		},
	}
	var supportedAuth []string
	for _, auth := range c.Spec.Authentication {
		if auth.StaticUsers != nil {
			containers := deploymentPatch.Spec.Template.Spec.Containers
			for i, container := range containers {
				if container.Name == "manager" {
					containers[i].Args = append(containers[i].Args,
						"--htpasswd-secret-name=$(HTPASSWD_SECRET_NAME)",
					)
					containers[i].Env = append(containers[i].Env,
						corev1.EnvVar{
							Name:  "HTPASSWD_SECRET_NAME",
							Value: auth.StaticUsers.HtpasswdSecret.Name,
						},
					)
					supportedAuth = append(supportedAuth, "Htpasswd")
				}
			}
		} else if auth.OIDC != nil {
			extraArgs := make([]string, 0)
			if len(auth.OIDC.RequiredClaims) > 0 {
				rclaims := make([]string, 0)
				for k, v := range auth.OIDC.RequiredClaims {
					rclaims = append(rclaims, fmt.Sprintf("%s=%s", k, v))
				}
				extraArgs = append(extraArgs, "--oidc-required-claim="+strings.Join(rclaims, ","))
			}

			containers := deploymentPatch.Spec.Template.Spec.Containers
			for i, container := range containers {
				if container.Name == "manager" {
					containers[i].Args = append(containers[i].Args, []string{
						"--oidc-issuer-url=$(API_SERVER_OIDC_ISSUER_URL)",
						"--oidc-client-id=$(API_SERVER_OIDC_CLIENT_ID)",
						"--oidc-ca-file=$(API_SERVER_OIDC_CA_FILE)",
						"--oidc-username-claim=$(API_SERVER_OIDC_USERNAME_CLAIM)",
						"--oidc-username-prefix=$(API_SERVER_OIDC_USERNAME_PREFIX)",
						"--oidc-groups-claim=$(API_SERVER_OIDC_GROUPS_CLAIM)",
						"--oidc-groups-prefix=$(API_SERVER_OIDC_GROUPS_PREFIX)",
						"--oidc-signing-algs=$(API_SERVER_OIDC_SIGNING_ALGS)",
					}...)
					containers[i].Args = append(containers[i].Args, extraArgs...)
					containers[i].Env = append(containers[i].Env, []corev1.EnvVar{
						{
							Name:  "API_SERVER_OIDC_ISSUER_URL",
							Value: auth.OIDC.IssuerURL,
						},
						{
							Name:  "API_SERVER_OIDC_CLIENT_ID",
							Value: auth.OIDC.ClientID,
						},
						{
							Name:  "API_SERVER_OIDC_CA_FILE",
							Value: "/run/oidc-certs/ca.crt",
						},
						{
							Name:  "API_SERVER_OIDC_USERNAME_CLAIM",
							Value: auth.OIDC.UsernameClaim,
						},
						{
							Name:  "API_SERVER_OIDC_USERNAME_PREFIX",
							Value: auth.OIDC.UsernamePrefix,
						},
						{
							Name:  "API_SERVER_OIDC_GROUPS_CLAIM",
							Value: auth.OIDC.GroupsClaim,
						},
						{
							Name:  "API_SERVER_OIDC_GROUPS_PREFIX",
							Value: auth.OIDC.GroupsPrefix,
						},
						{
							Name:  "API_SERVER_OIDC_SIGNING_ALGS",
							Value: strings.Join(auth.OIDC.SupportedSigningAlgs, ","),
						},
					}...)
					containers[i].VolumeMounts = append(containers[i].VolumeMounts, corev1.VolumeMount{
						MountPath: "/run/oidc-certs",
						ReadOnly:  true,
						Name:      "oidc-cert",
					})

					supportedAuth = append(supportedAuth, "OIDC")
				}
			}
			deploymentPatch.Spec.Template.Spec.Volumes = append(deploymentPatch.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: "oidc-cert",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: auth.OIDC.CertificateAuthority.Name,
					},
				},
			})
		} else if auth.ServiceAccount != nil {
			supportedAuth = append(supportedAuth, "Token")
		} else if auth.Anonymous != nil {
			supportedAuth = append(supportedAuth, "Anonymous")
		}
	}
	containers := deploymentPatch.Spec.Template.Spec.Containers
	for i, container := range containers {
		if container.Name == "manager" {
			for j, env := range containers[i].Env {
				if env.Name == AuthModeEnv {
					containers[i].Env[j].Value = strings.Join(supportedAuth, ",")
				}
			}
		}
	}

	managerEnvBytes, err := yaml.Marshal(deploymentPatch)
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
