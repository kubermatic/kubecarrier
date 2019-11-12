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

package spinner

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachSpinnerTo(t *testing.T) {
	msg := "Fake spinner message"
	tests := []struct {
		name    string
		f       func() error
		wantErr bool
	}{
		{
			"attach spinner to a function which returns no error",
			func() error {
				return nil
			},
			false,
		},
		{
			"attach spinner to a function which returns an error",
			func() error {
				return fmt.Errorf("this is an error")
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			s := wow.New(&b, spin.Get(spin.Dots), "")
			require.Equal(t, tt.wantErr, AttachSpinnerTo(s, msg, tt.f) != nil, "the error status is different")
			out := b.String()
			assert.Contains(t, msg, out, "the output should contain the initial message")
			if !tt.wantErr {
				assert.Contains(t, succeed, out, fmt.Sprintf("the output should contain %s", succeed))
			} else {
				assert.Contains(t, failed, out, fmt.Sprintf("the output should contain %s", failed))
			}
		})
	}
}
