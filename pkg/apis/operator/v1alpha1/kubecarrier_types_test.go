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

package v1alpha1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kubermatic/kubecarrier/pkg/apiserver/auth"
)

func TestKubeCarrierValidation(t *testing.T) {
	kubeCarrier := KubeCarrier{}
	err := kubeCarrier.Spec.API.Validate()
	assert.Error(t, err, "at least one authentication configuration should be specified")

	changed := kubeCarrier.Spec.API.Default()
	err = kubeCarrier.Spec.API.Validate()
	assert.Equal(t, true, changed)
	assert.NoError(t, err)

	kubeCarrier.Spec.API.Authentication = append(kubeCarrier.Spec.API.Authentication, AuthenticationConfig{Anonymous: &Anonymous{}})
	err = kubeCarrier.Spec.API.Validate()
	assert.Error(t, err, fmt.Sprintf("Duplicate %s configuration", auth.ProviderAnynymous))

	kubeCarrier.Spec.API.Authentication = Authentication{AuthenticationConfig{ServiceAccount: &ServiceAccount{}, Anonymous: &Anonymous{}}}
	err = kubeCarrier.Spec.API.Validate()
	assert.Error(t, err, "Authentication should have one and only one configuration")
}
