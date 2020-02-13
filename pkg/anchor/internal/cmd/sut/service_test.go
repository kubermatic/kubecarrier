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

package sut

import "testing"

func Test_k8sExpandEnvArg(t *testing.T) {
	type args struct {
		arg string
		env map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "no-replacement",
			args: args{
				arg: "test",
				env: map[string]string{},
			},
			want: "test",
		},
		{
			name: "simple-replacement",
			args: args{
				arg: "test-$(TEST_ID)",
				env: map[string]string{
					"TEST_ID": "1",
				},
			},
			want: "test-1",
		},
		{
			name: "escaped-replacement",
			args: args{
				arg: "test-$$(TEST_ID)-lala-$(TEST_ID)",
				env: map[string]string{
					"TEST_ID": "1",
				},
			},
			want: "test-$$(TEST_ID)-lala-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k8sExpandEnvArg(tt.args.arg, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("k8sExpandEnvArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("k8sExpandEnvArg() got = %v, want %v", got, tt.want)
			}
		})
	}
}
