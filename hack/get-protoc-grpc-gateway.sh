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

# Get the 'protoc' protocol buffer compiler
set -eu

DEST=tools
VERSION="1.14.3"

OS=$(uname | tr A-Z a-z)

GW_FILE="protoc-gen-grpc-gateway-v${VERSION}-${OS}-x86_64"
GW_URL="https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${VERSION}/${GW_FILE}"

SWAGGER_FILE="protoc-gen-swagger-v${VERSION}-${OS}-x86_64"
SWAGGER_URL="https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${VERSION}/${SWAGGER_FILE}"

SOURCE_URL="https://github.com/grpc-ecosystem/grpc-gateway/archive/v${VERSION}.zip"

mkdir -p $DEST/bin
curl -s --fail -L -# -o ${DEST}/bin/protoc-gen-grpc-gateway ${GW_URL}
curl -s --fail -L -# -o ${DEST}/bin/protoc-gen-swagger ${SWAGGER_URL}
curl -s --fail -L -# -o /tmp/protoc-gen-grpc-gateway.zip ${SOURCE_URL}
unzip -q -d /tmp/protoc-gen-grpc-gateway /tmp/protoc-gen-grpc-gateway.zip
mv /tmp/protoc-gen-grpc-gateway/grpc-gateway-${VERSION}/third_party ${DEST}/grpc-gateway-third_party
rm -rf /tmp/protoc-gen-grpc-gateway.zip /tmp/protoc-gen-grpc-gateway
chmod +x ${DEST}/bin/protoc-gen-grpc-gateway ${DEST}/bin/protoc-gen-swagger
