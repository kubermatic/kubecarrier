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

#!/usr/bin/env bash

set -eu

# As per official docs
# https://grpc-ecosystem.github.io/grpc-gateway/docs/usage.html
[[ -z $(whcih protoc-gen-grpc-gateway) ]] || go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
[[ -z $(whcih protoc-gen-swagger) ]] || go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
[[ -z $(whcih protoc-gen-go) ]] || go get -u github.com/golang/protobuf/protoc-gen-go

# Add protoc and protoc-gen-go tools to PATH
export PATH=$PWD/bin:$PATH
PROJECT=$PWD
GOPATH=$(go env GOPATH)

PBUFS=(
  pkg/apis/apiserver/v1alpha1
)

# change into each protobuf directory
for pkg in ${PBUFS} ; do
  abs_path=${PROJECT}/${pkg}
  echo Generating from '*.proto' in $abs_path
  protoc \
    --go_out=plugins=grpc:${abs_path}  \
    --grpc-gateway_out=logtostderr=true:${abs_path} \
    --swagger_out=logtostderr=true:${abs_path} \
    -I${GOPATH}/src \
    -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -I=${abs_path} \
    $(find ${abs_path} -type f -name '*.proto')

  for x in $(find ${abs_path} -type f -name '*pb*.go'); do
    echo $x
    cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - $x > $x.tmp
    mv $x.tmp $x
  done
done
