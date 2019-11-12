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

package operator

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

type kustomizeContextMock struct {
	mock.Mock
}

var _ kustomize.KustomizeContext = (*kustomizeContextMock)(nil)

func (k *kustomizeContextMock) ReadFile(path string) ([]byte, error) {
	args := k.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (k *kustomizeContextMock) WriteFile(path string, content []byte) error {
	args := k.Called(path, content)
	return args.Error(0)
}

func (k *kustomizeContextMock) Build(path string) ([]unstructured.Unstructured, error) {
	args := k.Called(path)
	return args.Get(0).([]unstructured.Unstructured), args.Error(1)
}

type kustomizeStub struct {
	*kustomizeContextMock
}

func (k *kustomizeStub) ForHTTP(fs http.FileSystem) kustomize.KustomizeContext {
	return k
}

func TestManifests(t *testing.T) {
	c := Config{
		Namespace: "test3000",
	}
	m := &kustomizeContextMock{}
	m.On("ReadFile", "/default/kustomization.yaml").Return([]byte("namespace: default"), nil)
	m.On("WriteFile", "/default/kustomization.yaml", mock.Anything).Return(nil)
	m.On("Build", mock.Anything).Return([]unstructured.Unstructured{}, nil)

	_, err := Manifests(&kustomizeStub{m}, c)
	require.NoError(t, err, "unexpected error")

	m.AssertCalled(t, "WriteFile", "/default/kustomization.yaml", []byte(fmt.Sprintf(`images:
- name: quay.io/kubecarrier/operator
  newTag: %s
namespace: test3000
`, version.Get().Version)))
}
