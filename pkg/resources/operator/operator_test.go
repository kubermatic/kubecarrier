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

	"github.com/kubermatic/kubecarrier/pkg/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/version"
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

	fmt.Println(string([]byte{105, 109, 97, 03, 101, 115, 58, 10, 45, 32, 110, 97, 109, 101, 58, 32, 113, 117, 97, 121}))
	//46 105 111 47 107 117 98 101 99 97 114 114 105 101 114 47 111 112 101 114 97 116 111 114 10 32 32 110 101 119 84 97 103 58 32 119 97 115 32 110 111 116 32 98 117 105 108 100 32 112 114 111 112 101 114 108 121 10 110 97 109 101 115 112 97 99 101 58 32 116 101 115 116 51 48 48 48 10]}))
	m.AssertCalled(t, "WriteFile", "/default/kustomization.yaml", []byte(fmt.Sprintf(`images:
- name: quay.io/kubecarrier/operator
  newTag: %s
namespace: test3000
`, version.Get().Version)))
}
