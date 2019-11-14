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
	"fmt"
	"runtime"
	"strconv"
	"time"
)

// Values are provided by compile time -ldflags.
var (
	empty     = "was not build properly"
	Version   = empty
	Branch    = empty
	Commit    = empty
	BuildDate = empty
)

// Info contains build information supplied during compile time.
type Info struct {
	Version   string    `json:"version"`
	Branch    string    `json:"branch"`
	Commit    string    `json:"commit"`
	BuildDate time.Time `json:"buildTime"`
	GoVersion string    `json:"goVersion"`
	Platform  string    `json:"platform"`
}

// Get returns the build-in version and platform information
func Get() Info {
	v := Info{
		Version:   Version,
		Branch:    Branch,
		Commit:    Commit,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	if BuildDate != empty {
		i, err := strconv.ParseInt(BuildDate, 10, 64)
		if err != nil {
			panic(fmt.Errorf("error parsing build time: %v", err))
		}
		v.BuildDate = time.Unix(i, 0).UTC()
	}

	return v
}
