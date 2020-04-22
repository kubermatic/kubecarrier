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
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/kubermatic/kubecarrier/pkg/testutil"
)

func TestManifests(t *testing.T) {
	const (
		goldenFile = "api-server.golden.yaml"
	)
	c := Config{
		Namespace: "kubecarrier-system-10",
		Name:      "foo",
	}

	manifests, err := Manifests(c)
	require.NoError(t, err, "unexpected error")
	yManifest, err := yaml.Marshal(manifests)
	require.NoError(t, err, "cannot marshall given manifests")

	if _, present := os.LookupEnv(testutil.OverrideGoldenEnv); present {
		require.NoError(t, ioutil.WriteFile(goldenFile, yManifest, 0640))
	}

	yGoldenManifest, err := ioutil.ReadFile(goldenFile)
	require.NoError(t, err)
	if string(yManifest) != string(yGoldenManifest) {
		t.Logf("generated manifests differ from the golden file:\n%s", cmp.Diff(
			string(yGoldenManifest), string(yManifest)))
		t.FailNow()
	}
}
