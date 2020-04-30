#!/usr/bin/env bash

# Copyright 2019 The KubeCarrier Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

command -v protoc >/dev/null || (
  curl -sL --output /tmp/protoc.zip https://github.com/google/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && unzip /tmp/protoc.zip -d /usr && rm /tmp/protoc.zip
)
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${PROTOC_GATEWAY_VERSION}
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${PROTOC_GATEWAY_VERSION}
go get -u github.com/golang/protobuf/protoc-gen-go@v${PROTOC_GEN_GO_VERSION}
