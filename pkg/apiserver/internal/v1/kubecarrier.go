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

package v1

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/golang/protobuf/ptypes/timestamp"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/oidc"
	"github.com/kubermatic/kubecarrier/pkg/internal/version"
)

type KubeCarrierServer struct{}

var _ v1.KubeCarrierServer = (*KubeCarrierServer)(nil)

func (v KubeCarrierServer) Version(context.Context, *v1.VersionRequest) (*v1.APIVersion, error) {
	versionInfo := version.Get()
	return &v1.APIVersion{
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

func (v KubeCarrierServer) WhoAmI(ctx context.Context, _ *empty.Empty) (*v1.UserInfo, error) {
	user, present := oidc.ExtractUserInfo(ctx)
	if !present {
		return nil, fmt.Errorf("unauthorized")
	}
	return &v1.UserInfo{
		User:   user.GetName(),
		Groups: user.GetGroups(),
	}, nil
}
