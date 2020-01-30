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

package catapult

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/go-playground/validator/v10"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/image"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier controller manager setup.
type Config struct {
	// Name is the name of the Catapult instance.
	Name string `validate:"required"`

	// ProviderNamespace that the Catapult instance should be deployed into.
	ProviderNamespace string `validate:"required"`

	// KubecarrierNamespace where kubecarrier system components are deployed
	KubecarrierNamespace string `validate:"required"`

	// KubeconfigSecretName of the secret holding the service cluster kubeconfig under the ${KubeconfigSecretKey} key
	KubeconfigSecretName string `validate:"required"`

	// KubeconfigSecretKey key of the secret holding the service cluster kubeconfig
	KubeconfigSecretKey string `validate:"required"`

	operatorv1alpha1.CatapultMappingSpec `validate:"required,dive,required"`
}

var k = kustomize.NewDefaultKustomize()
var validate = validator.New()

func Manifests(c Config) ([]unstructured.Unstructured, error) {
	if err := validate.Struct(c); err != nil {
		return nil, err
	}

	v := version.Get()
	kc := k.ForHTTP(vfs)
	const envPathFile = "manager_env_patch.yaml"
	if err := kc.MkLayer("man", types.Kustomization{
		Namespace:  c.ProviderNamespace,
		NamePrefix: fmt.Sprintf("%s-", c.Name),
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/catapult",
				NewTag: v.Version,
			},
		},
		PatchesStrategicMerge: []types.PatchStrategicMerge{
			envPathFile,
		},
		Resources: []string{"../default"},
	}); err != nil {
		return nil, fmt.Errorf("cannot MkLayer: %w", err)
	}

	t := template.Must(template.New("").Parse(`
apiVersion: apps/v1
kind: Deployment
metadata:
  name: manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        env:
        - name: KUBECARRIER_NAMESPACE
          value: "{{ .KubecarrierNamespace }}"
        - name: MASTER_GROUP
          value: "{{ .MasterGroup }}"
        - name: MASTER_KIND
          value: "{{ .MasterKind }}"
        - name: SERVICE_GROUP
          value: "{{ .ServiceGroup }}"
        - name: SERVICE_KIND
          value: "{{ .ServiceKind }}"
        - name: RESOURCE_VERSION
          value: "{{ .ObjectVersion }}"
        volumeMounts:
        - mountPath: /kubeconfig
          name: kubeconfig
      volumes:
        - name: kubeconfig
          secret:
            optional: false
            secretName: "{{ .KubeconfigSecretName }}"
            items:
              - key: "{{ .KubeconfigSecretKey }}"
                path: "kubeconfig"
`))

	{
		b := new(bytes.Buffer)
		if err := t.Execute(b, c); err != nil {
			return nil, fmt.Errorf("templating strategic patch merge")
		}
		if err := kc.WriteFile("/man/"+envPathFile, b.Bytes()); err != nil {
			return nil, fmt.Errorf("writing to file")
		}
	}

	// execute kustomize
	objects, err := kc.Build("/man")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
