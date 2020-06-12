/*
Copyright 2020 The KubeCarrier Authors.

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

package v1

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

func TestValidateGetRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *v1.GetRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &v1.GetRequest{
				Name:    "test-name",
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "missing name",
			req: &v1.GetRequest{
				Account: "test-namespace",
			},
			expectedError: fmt.Errorf("missing name"),
		},
		{
			name: "valid request",
			req: &v1.GetRequest{
				Name:    "test-name",
				Account: "test-namespace",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, validateGetRequest(test.req))
		})
	}
}

func TestValidateListRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *v1.ListRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &v1.ListRequest{
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "invalid limit",
			req: &v1.ListRequest{
				Account: "test-namespace",
				Limit:   -1,
			},
			expectedError: fmt.Errorf("invalid limit: should not be negative number"),
		},
		{
			name: "invalid label selector",
			req: &v1.ListRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=====name1",
			},
			expectedError: fmt.Errorf("invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
		},
		{
			name: "valid request",
			req: &v1.ListRequest{
				Account: "test-namespace",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateListRequest(test.req)
			if err == nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestValidateWatchRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *v1.WatchRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &v1.WatchRequest{
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "invalid label selector",
			req: &v1.WatchRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=====name1",
			},
			expectedError: fmt.Errorf("invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
		},
		{
			name: "valid request",
			req: &v1.WatchRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label==name1",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validateWatchRequest(test.req)
			if err == nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}
