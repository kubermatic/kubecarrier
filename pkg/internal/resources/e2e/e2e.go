/*
Copyright 2019 The Kubecarrier Authors.

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

package e2e

import (
	"fmt"
	"net/http"

	"gopkg.in/yaml.v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"

	statikfs "github.com/rakyll/statik/fs"
)

// Config holds the config information to generate the kubecarrier operator setup.
type Config struct {
	// Namespace the e2e operator should be deployed into.
	Namespace string
}

type kustomizeFactory interface {
	ForHTTP(fs http.FileSystem) kustomize.KustomizeContext
}

func Manifests(k kustomizeFactory, c Config) ([]unstructured.Unstructured, error) {
	vfs, err := statikfs.New()
	if err != nil {
		return nil, fmt.Errorf("cannot get statik fs: %w", err)
	}
	kc := k.ForHTTP(vfs)

	// patch settings
	kustomizePath := "/e2e/default/kustomization.yaml"
	kustomizeBytes, err := kc.ReadFile(kustomizePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", kustomizePath, err)
	}
	kmap := map[string]interface{}{}
	if err := yaml.Unmarshal(kustomizeBytes, &kmap); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", kustomizePath, err)
	}

	// patch namespace
	kmap["namespace"] = c.Namespace
	// patch image tag
	v := version.Get()
	kmap["images"] = []map[string]string{
		{
			"name":   "quay.io/kubecarrier/e2e",
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

	// execute kustomize
	objects, err := kc.Build("/e2e/default")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
