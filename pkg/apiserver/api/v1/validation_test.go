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
)

func TestValidateGetRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *GetRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &GetRequest{
				Name:    "test-name",
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "missing name",
			req: &GetRequest{
				Account: "test-namespace",
			},
			expectedError: fmt.Errorf("missing name"),
		},
		{
			name: "valid request",
			req: &GetRequest{
				Name:    "test-name",
				Account: "test-namespace",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedError, test.req.Validate())
		})
	}
}

func TestValidateListRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *ListRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &ListRequest{
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "invalid limit",
			req: &ListRequest{
				Account: "test-namespace",
				Limit:   -1,
			},
			expectedError: fmt.Errorf("invalid limit: should not be negative number"),
		},
		{
			name: "invalid label selector",
			req: &ListRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=====name1",
			},
			expectedError: fmt.Errorf("invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
		},
		{
			name: "valid request",
			req: &ListRequest{
				Account: "test-namespace",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.req.Validate()
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
		req           *WatchRequest
		expectedError error
	}{
		{
			name: "missing namespace",
			req: &WatchRequest{
				Account: "",
			},
			expectedError: fmt.Errorf("missing namespace"),
		},
		{
			name: "invalid label selector",
			req: &WatchRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label=====name1",
			},
			expectedError: fmt.Errorf("invalid LabelSelector: unable to parse requirement: found '==', expected: identifier"),
		},
		{
			name: "valid request",
			req: &WatchRequest{
				Account:       "test-namespace",
				LabelSelector: "test-label==name1",
			},
			expectedError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.req.Validate()
			if err == nil {
				assert.Equal(t, test.expectedError, err)
			} else {
				assert.EqualError(t, err, test.expectedError.Error())
			}
		})
	}
}
