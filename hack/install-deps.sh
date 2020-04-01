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

PROTOC_VERSION="3.11.4"
PROTOC_GATEWAY_VERSION="1.14.3"

function install_protoc() {
  local DEST=$1
  local VERSION=${PROTOC_VERSION}

  local OS=$(uname | tr A-Z a-z)
  if [[ $OS == 'darwin' ]]; then
    OS=osx     # protoc names downloads with OSX, not darwin
  fi

  local FILE="protoc-${VERSION}-${OS}-x86_64.zip"
  local URL="https://github.com/google/protobuf/releases/download/v${VERSION}/${FILE}"

  mkdir -p $DEST
  curl --fail -L -# -o protoc.zip ${URL}
  unzip -d ${DEST}/protoc-bin protoc.zip
  chmod +x ${DEST}/protoc-bin/bin/protoc
  ln -sf protoc-bin/bin/protoc ${DEST}/protoc
  rm -f protoc.zip
}

export GOBIN=${1}/bin
install_protoc $1/bin
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.14.3
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.14.3
go get -u github.com/golang/protobuf/protoc-gen-go@v1.3.5
