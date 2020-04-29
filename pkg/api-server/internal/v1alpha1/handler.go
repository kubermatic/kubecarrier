/*
Copyright 2020 The KubeCarrier Authors.

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

package v1alpha1

import (
	"context"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/kubermatic/kubecarrier/pkg/api-server/api/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

type KubeCarrierServer struct{}

var _ v1alpha1.KubeCarrierServer = (*KubeCarrierServer)(nil)

func (v KubeCarrierServer) Version(context.Context, *v1alpha1.VersionRequest) (*v1alpha1.APIVersion, error) {
	versionInfo := version.Get()
	return &v1alpha1.APIVersion{
		Version: versionInfo.Version,
		Branch:  versionInfo.Branch,
		BuildDate: &timestamp.Timestamp{
			Seconds: versionInfo.BuildDate.Unix(),
			Nanos:   int32(versionInfo.BuildDate.Nanosecond()),
		},
		GoVersion: versionInfo.GoVersion,
		Platform:  versionInfo.Platform,
	}, nil
}
