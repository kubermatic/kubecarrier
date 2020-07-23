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

# Development Tooling Container
# build by running `make build-image-dev`

FROM golang:1.14.2

RUN apt-get -qq update && \
  apt-get -qqy install ed curl gettext zip python3 python3-pip git jq make && \
  rm -rf /var/lib/apt/lists/* && \
  pip3 install pre-commit yq


# Allowed to use path@version
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV PATH=${PATH}:/usr/local/go/bin:/usr/local/protoc/bin:${GOPATH}/bin

# versions without the `v` prefix
ARG GOLANGCI_LINT_VERSION
ARG STATIK_VERSION
ARG CONTROLLER_GEN_VERSION
ARG PROTOC_VERSION
ARG PROTOC_GEN_GO_VERSION
ARG PROTOC_GRPC_GATEWAY_VERSION
ARG TESTIFY_VERSION

RUN echo $PATH && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v${GOLANGCI_LINT_VERSION} && \
  go get golang.org/x/tools/cmd/goimports && \
  go get github.com/pablo-ruth/go-init && \
  go get github.com/stretchr/testify@v${TESTIFY_VERSION} && \
  go get github.com/thetechnick/statik@v${STATIK_VERSION} && \
  go get sigs.k8s.io/controller-tools/cmd/controller-gen@v${CONTROLLER_GEN_VERSION} && \
  go get -u github.com/golang/protobuf/protoc-gen-go@v${PROTOC_GEN_GO_VERSION} && \
  go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${PROTOC_GRPC_GATEWAY_VERSION} && \
  go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${PROTOC_GRPC_GATEWAY_VERSION} && \
  curl -sL --output /tmp/protoc.zip https://github.com/google/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
  mkdir -p /usr/local/protoc && \
  unzip /tmp/protoc.zip -d /usr/local/protoc && \
  rm /tmp/protoc.zip

RUN curl -sL --output /tmp/boilerplate.zip https://github.com/kubermatic-labs/boilerplate/releases/download/v0.1.1/boilerplate_0.1.1_linux_amd64.zip && unzip /tmp/boilerplate.zip -d /usr/local/bin && rm -Rf /tmp/boilerplate.zip
