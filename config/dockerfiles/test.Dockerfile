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

# Should be build from the Makefile, `make build-image-test`

FROM ubuntu:18.04

RUN apt-get -qq update && apt-get -qqy install \
  apt-transport-https \
  build-essential \
  ca-certificates \
  curl \
  ed \
  gettext \
  git \
  gnupg2 \
  jq \
  python3-pip \
  software-properties-common \
  zip \
  && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://get.docker.com | sh
RUN curl -sL --output /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.16.0/bin/linux/amd64/kubectl && chmod a+x /usr/local/bin/kubectl
RUN curl -sL https://get.helm.sh/helm-v3.1.2-linux-amd64.tar.gz | tar -C /tmp -xz && mv /tmp/linux-amd64/helm /usr/bin/helm && rm -Rf /tmp/linux-amd64 && helm repo add stable https://charts.helm.sh/stable && helm repo update
RUN curl -sL https://dl.google.com/go/go1.14.linux-amd64.tar.gz | tar -C /usr/local -xz
RUN curl -sL --output /tmp/boilerplate.zip https://github.com/kubermatic-labs/boilerplate/releases/download/v0.1.1/boilerplate_0.1.1_linux_amd64.zip && unzip /tmp/boilerplate.zip -d /usr/local/bin && rm -Rf /tmp/boilerplate.zip
ENV PATH=${PATH}:/usr/local/go/bin:/root/go/bin
# This LC_ALL is needed for yq. https://stackoverflow.com/questions/18649512/unicodedecodeerror-ascii-codec-cant-decode-byte-0xe2-in-position-13-ordinal
ENV LC_ALL=C.UTF-8
# Allowed to use path@version syntax to install controller-gen
ENV GO111MODULE=on
COPY lib.sh .
COPY install-deps.sh .
RUN ./install-deps.sh
RUN go env

# binary will be $(go env GOPATH)/bin/golangci-lint
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.24.0

RUN curl -sL --output /usr/local/bin/kind https://github.com/kubernetes-sigs/kind/releases/download/v0.8.1/kind-linux-amd64 && chmod a+x /usr/local/bin/kind
RUN curl -sL https://github.com/kyoh86/richgo/releases/download/v0.3.3/richgo_0.3.3_linux_amd64.tar.gz | tar -xz -C /tmp/ && mv /tmp/richgo /usr/local/bin

ARG kubebuilder_version=2.1.0
RUN curl -sL https://go.kubebuilder.io/dl/${kubebuilder_version}/linux/amd64 | tar -xz -C /tmp/ && mv /tmp/kubebuilder_${kubebuilder_version}_linux_amd64 /usr/local/kubebuilder

RUN go get golang.org/x/tools/cmd/goimports
RUN go get github.com/rakyll/statik
RUN go get github.com/pablo-ruth/go-init
# Install controller-gen in the dockerfile, otherwise it will be installed during `make generate` which will modify the go.mod
# and go.sum files, and make the directory dirty.
RUN go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.9
RUN pip3 install pre-commit awscli yq

WORKDIR /src

# Create pre-commit cache, that is download required pre-commit repos
COPY .pre-commit-config.yaml .pre-commit-config.yaml
RUN git init && (pre-commit run || true) && rm -Rvf .git

COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY start-docker.sh /usr/local/bin/start-docker.sh

VOLUME /var/lib/docker
