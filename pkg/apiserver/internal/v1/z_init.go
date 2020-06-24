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

package v1

import (
	"net/http"

	statikfs "github.com/thetechnick/statik/fs"
)

// vfs is a virtual file system to access the operator config.
var vfs http.FileSystem

// don't rename this file!
// this init() function must be called after statik.go
func init() {
	var err error
	vfs, err = statikfs.New()
	if err != nil {
		panic(err)
	}
}
