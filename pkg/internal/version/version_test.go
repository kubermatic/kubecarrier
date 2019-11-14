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
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	assert.Equal(t, Version, empty)
	assert.Equal(t, Branch, empty)
	assert.Equal(t, Commit, empty)
	assert.Equal(t, BuildDate, empty)

}

func TestGet(t *testing.T) {
	Version = "1.2.3"
	Branch = "branch"
	Commit = "commit"
	BuildDate = "1573126751"

	v := Get()
	assert.Equal(t, Version, v.Version)
	assert.Equal(t, Branch, v.Branch)
	assert.Equal(t, Commit, v.Commit)
	assert.Equal(t, BuildDate, strconv.Itoa(int(v.BuildDate.Unix())))
}
