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
	"io"
	"os"

	statikfs "github.com/rakyll/statik/fs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

// Config holds the config information to generate the kubecarrier controller manager setup.
type Config struct {
	// Namespace is the kubecarrier controller manager should be deployed into.
	Namespace string
}

var k = kustomize.NewDefaultKustomize()

func Manifests(c Config) ([]unstructured.Unstructured, error) {
	kustomizeFs := fs.MakeFsInMemory()
	if err := statikfs.Walk(vfs, "/", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return kustomizeFs.Mkdir(path)
		}
		Fout, err := kustomizeFs.Create(path)
		if err != nil {
			return err
		}
		Fin, err := vfs.Open(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(Fout, Fin); err != nil {
			return err
		}
		if err := Fout.Close(); err != nil {
			return err
		}
		if err := Fin.Close(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("fs creation: %w", err)
	}
	kc := k.For(kustomizeFs)

	// patch settings
	kustomizePath := "/default/kustomization.yaml"
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
			"name":   "quay.io/kubecarrier/manager",
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
	objects, err := kc.Build("/default")
	if err != nil {
		return nil, fmt.Errorf("running kustomize build: %w", err)
	}
	return objects, nil
}
