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

package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func TestGenerateWebhookPath(t *testing.T) {
	testWebhookScheme := runtime.NewScheme()
	if err := catalogv1alpha1.AddToScheme(testWebhookScheme); err != nil {
		panic(err)
	}
	catalogEntry := &catalogv1alpha1.CatalogEntry{}

	path, err := GenerateMutateWebhookPath(catalogEntry, testWebhookScheme)
	require.NoError(t, err)
	assert.Equal(t, path, "/mutate-catalog-kubecarrier-io-v1alpha1-catalogentry")

	path, err = GenerateValidateWebhookPath(catalogEntry, testWebhookScheme)
	require.NoError(t, err)
	assert.Equal(t, path, "/validate-catalog-kubecarrier-io-v1alpha1-catalogentry")
}

func TestIsDNS1123Label(t *testing.T) {
	tests := []struct {
		s              string
		isDNS1123Label bool
	}{
		{
			"example.cloud",
			false,
		},
		{
			"Example-cloud",
			false,
		},
		{
			"-example-cloud",
			false,
		},
		{
			"example-cloud-",
			false,
		},
		{
			"examp1e-cloud",
			true,
		},
		{
			"example-cloud",
			true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.isDNS1123Label, IsDNS1123Label(test.s))
	}
}
