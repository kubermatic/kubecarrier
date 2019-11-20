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

package resources

//go:generate bash -c "statik -src=../../../config/operator -p operator -f -c ''"
//go:generate bash -c "statik -src=../../../config/manager -p manager -f -c ''"
//go:generate bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/2019/ | cat - operator/statik.go > operator/statik.go.tmp; mv operator/statik.go.tmp operator/statik.go"
//go:generate bash -c "cat ../../../hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/2019/ | cat - manager/statik.go > manager/statik.go.tmp; mv manager/statik.go.tmp manager/statik.go"
