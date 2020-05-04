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

DEST="tools/protoc"
VERSION="3.11.4"

OS=$(uname | tr A-Z a-z)
if [[ $OS == 'darwin' ]]; then
  OS=osx     # protoc names downloads with OSX, not darwin
fi

FILE="protoc-${VERSION}-${OS}-x86_64.zip"
URL="https://github.com/google/protobuf/releases/download/v${VERSION}/${FILE}"

mkdir -p $DEST
curl --fail -s -L -# -o protoc.zip ${URL}
unzip -q -d ${DEST} protoc.zip
chmod +x ${DEST}/bin/protoc
ln -s protoc-bin/bin/protoc ${DEST}/protoc
rm -f protoc.zip
