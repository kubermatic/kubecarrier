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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestLoggerOutput(t *testing.T) {
	l := NewLogger(t)

	g := l.WithName("a").WithName("b").WithValues("c", "d")
	g.Info("msg")
	require.Equal(t, []string{"a", "b"}, g.(*Logger).names)
	assert.Equal(t, "d", g.(*Logger).values["c"])
}
