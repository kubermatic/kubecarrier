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

package main

import (
	"testing"
)

func Test_fileExtension(t *testing.T) {
	tests := []struct {
		name, file, extension string
	}{
		{
			name:      "Simple extension",
			file:      "e2e-test.go",
			extension: "go",
		},
		{
			name:      "dot in between",
			file:      "script.e2e-test.sh",
			extension: "sh",
		},
		{
			name:      "no extension",
			file:      "script",
			extension: "",
		},
		{
			name:      "dot in folder structure",
			file:      "k8s.io/script",
			extension: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ext := fileExtension(test.file)
			if ext != test.extension {
				t.Errorf("file extension should be %q, is: %q", test.extension, ext)
			}
		})
	}
}
