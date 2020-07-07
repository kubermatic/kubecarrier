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

package version

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubermatic/kubecarrier/pkg/internal/version"
	"github.com/kubermatic/utils/pkg/testutil"
)

func TestNewCommand(t *testing.T) {
	// Override the version.Version property,
	// to give a nice test output representation
	version.Version = "v1.1.0"

	t.Run("default output", func(t *testing.T) {
		log := testutil.NewLogger(t)
		var (
			stdout, stderr bytes.Buffer
		)
		cmd := NewCommand(log)
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)

		err := cmd.Execute()
		require.NoError(t, err, "execute returned an error")

		assert.Equal(t, "v1.1.0\n", stdout.String())
		assert.Equal(t, "", stderr.String())
	})

	t.Run("long output", func(t *testing.T) {
		log := testutil.NewLogger(t)
		var (
			stdout, stderr bytes.Buffer
		)
		cmd := NewCommand(log)
		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.SetArgs([]string{"--full"})
		v := version.Get()

		err := cmd.Execute()
		require.NoError(t, err, "execute returned an error")

		assert.Equal(t, fmt.Sprintf(`branch: was not build properly
buildTime: "0001-01-01T00:00:00Z"
commit: was not build properly
goVersion: %s
platform: %s
version: v1.1.0
`, v.GoVersion, v.Platform), stdout.String())
	})
}
