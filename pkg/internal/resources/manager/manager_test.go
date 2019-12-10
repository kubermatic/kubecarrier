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
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/internal/kustomize"
	"github.com/kubermatic/kubecarrier/pkg/testutil"
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
	const (
		goldenFile = "manager.golden"
	)
	c := Config{
		Namespace: "test3000",
	}

	manifests, err := Manifests(c)
	require.NoError(t, err, "unexpected error")
	yManifest, err := yaml.Marshal(manifests)
	require.NoError(t, err, "cannot marshall given manifests")

	if _, present := os.LookupEnv(testutil.OverrideGoldenEnv); present {
		require.NoError(t, ioutil.WriteFile(goldenFile, yManifest, 0440))
	}
	yGoldenManifest, err := ioutil.ReadFile(goldenFile)
	require.NoError(t, err)
	if string(yManifest) != string(yGoldenManifest) {
		t.Logf("generated manifests differ from the golden file:\n%s", cmp.Diff(
			string(yGoldenManifest), string(yManifest)))
		t.FailNow()
	}
}
