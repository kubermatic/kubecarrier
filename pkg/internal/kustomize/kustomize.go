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

package kustomize

import (
	"fmt"
	"io"
	"net/http"
	"os"

	statikfs "github.com/rakyll/statik/fs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/v3/k8sdeps/transformer"
	"sigs.k8s.io/kustomize/v3/k8sdeps/validator"
	"sigs.k8s.io/kustomize/v3/pkg/fs"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/loader"
	"sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/target"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

// Kustomize holds factories and config options.
type Kustomize struct {
	unstructuredFactory ifc.KunstructuredFactory
	patchFactory        resmap.PatchFactory
	resmapFactory       *resmap.Factory
	pluginLoader        *plugins.Loader
	validator           ifc.Validator
}

// NewDefaultKustomize creates a Kustomize instance with a sane default config.
func NewDefaultKustomize() *Kustomize {
	pluginConfig := plugins.DefaultPluginConfig()

	k := &Kustomize{
		unstructuredFactory: kunstruct.NewKunstructuredFactoryImpl(),
		patchFactory:        transformer.NewFactoryImpl(),
		validator:           validator.NewKustValidator(),
	}
	k.resmapFactory = resmap.NewFactory(
		resource.NewFactory(k.unstructuredFactory),
		k.patchFactory,
	)
	k.pluginLoader = plugins.NewLoader(
		pluginConfig,
		k.resmapFactory,
	)
	return k
}

// KustomizeContext combines a Kustomize instance with FileSystem operations.
type KustomizeContext interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	MkLayer(name string, kustomization types.Kustomization) error
	Build(path string) ([]unstructured.Unstructured, error)
}

// For returns a new KustomizeContext using the given FileSystem.
func (k *Kustomize) For(fs fs.FileSystem) KustomizeContext {
	return &kustomizeContext{
		Kustomize: k,
		fs:        fs,
	}
}

// ForHTTP returns a new KustomizeContext using the given http.FileSystem.
func (k *Kustomize) ForHTTP(httpFS http.FileSystem) KustomizeContext {
	kustomizeFs := fs.MakeFsInMemory()
	if err := statikfs.Walk(httpFS, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return kustomizeFs.Mkdir(path)
		}

		// Read data. It should not error since it's in memory FS
		Fin, err := httpFS.Open(path)
		if err != nil {
			return err
		}
		// Write modified data. It should not error since it's in memory FS
		Fout, err := kustomizeFs.Create(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(Fout, Fin); err != nil {
			return fmt.Errorf("cannot copy data: %w", err)
		}

		if err := Fin.Close(); err != nil {
			return err
		}
		if err := Fout.Close(); err != nil {
			return err
		}

		return nil
	}); err != nil {
		panic(err)
	}

	return &kustomizeContext{
		Kustomize: k,
		fs:        kustomizeFs,
	}
}

// kustomizeContext combines a Kustomize instance with a FileSystem to operate on.
type kustomizeContext struct {
	*Kustomize
	fs fs.FileSystem
}

var _ KustomizeContext = (*kustomizeContext)(nil)

// ReadFile reads the file's content from the underlying FS.
func (k *kustomizeContext) ReadFile(path string) ([]byte, error) {
	return k.fs.ReadFile(path)
}

// MkLayer makes directory in the underlying FS, and initialize its kustomization.yaml
func (k *kustomizeContext) MkLayer(name string, kustomization types.Kustomization) error {
	if err := k.fs.Mkdir(name); err != nil {
		return fmt.Errorf("cannot create dir: %w", err)
	}

	kustomizationBytes, err := yaml.Marshal(kustomization)
	if err != nil {
		return fmt.Errorf("cannot yaml marshal: %w", err)
	}

	if err := k.WriteFile("/"+name+"/kustomization.yaml", kustomizationBytes); err != nil {
		return fmt.Errorf("cannot write marshal: %w", err)
	}
	return nil
}

// WriteFile writes the given content to the underlying FS.
func (k *kustomizeContext) WriteFile(path string, content []byte) error {
	return k.fs.WriteFile(path, content)
}

// Build is equivalent to running `kustomize build path` for the given filesystem.
func (k *kustomizeContext) Build(path string) ([]unstructured.Unstructured, error) {
	ldr, err := loader.NewLoader(loader.RestrictionRootOnly, k.validator, path, k.fs)
	if err != nil {
		return nil, fmt.Errorf("creating loader: %w", err)
	}

	kt, err := target.NewKustTarget(ldr, k.resmapFactory, k.patchFactory, k.pluginLoader)
	if err != nil {
		return nil, fmt.Errorf("creating kustomize target: %w", err)
	}

	m, err := kt.MakeCustomizedResMap()
	if err != nil {
		return nil, fmt.Errorf("creating res map: %w", err)
	}

	var objects []unstructured.Unstructured
	for _, res := range m.Resources() {
		adapter, ok := res.Kunstructured.(*kunstruct.UnstructAdapter)
		if !ok {
			return nil, fmt.Errorf("cannot convert kustomize item instance to Unstructured")
		}

		objects = append(objects, adapter.Unstructured)
	}
	return objects, nil
}
