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

package operator

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/image"
	"sigs.k8s.io/kustomize/v3/pkg/types"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier controller manager setup.
type Config struct {
	// Namespace is the kubecarrier controller manager should be deployed into.
	Namespace string
}

var k = kustomize.NewDefaultKustomize()

// Manifests generate all required manifests for the Operator
// See https://github.com/kubermatic/kubecarrier/issues/95 for discussion
func Manifests(c Config) ([]unstructured.Unstructured, error) {
	v := version.Get()
	kc := k.ForHTTP(vfs)
	if err := kc.MkLayer("man", types.Kustomization{
		Namespace: c.Namespace,
		Images: []image.Image{
			{
				Name:   "quay.io/kubecarrier/operator",
				NewTag: v.Version,
			},
		},
		Resources: []string{"../default"},
	}); err != nil {
		return nil, fmt.Errorf("cannot mkdir: %w", err)
	}

	// execute kustomize
	objects, err := kc.Build("/man")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
