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

package resources

import (
	"strings"

	"github.com/kubermatic/kubecarrier/pkg/version"
)

//go:generate bash -c "statik -src=../../config/operator -p operator -f -c \"$DOLLAR(cat ../../hack/boilerplate/boilerplate.go.txt | sed s/YEAR/2019/g )\""

// default images
const (
	KubecarrierRegistry = "quay.io/kubermatic/kubecarrier"
	operatorRepository  = KubecarrierRegistry + "/operator"
)

var (
	DefaultImageTag string
)

func init() {
	v := version.Get()
	tag := v.Version
	DefaultImageTag = strings.Replace(tag, "+", "_", -1)
}
