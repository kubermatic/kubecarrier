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

package manager

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubermatic/kubecarrier/pkg/internal/resources/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier controller manager setup.
type Config struct {
	// Namespace is the kubecarrier controller manager should be deployed into.
	Namespace string
}

var k = kustomize.NewDefaultKustomize()

func Manifests(c Config) ([]unstructured.Unstructured, error) {
	v := version.Get()
	kc, err := k.ForHTTPWithReplacement(vfs, map[string]string{
		"kubecarrier-system":                  c.Namespace,
		"quay.io/kubecarrier/manager:lastest": "quay.io/kubecarrier/manager:" + v.Version,
	})
	if err != nil {
		return nil, err
	}

	// execute kustomize
	objects, err := kc.Build("/default")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
