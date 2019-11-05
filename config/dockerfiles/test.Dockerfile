# Copyright 2019 The Kubecarrier authors.
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

FROM golang:1.13

RUN apt-get update && apt-get -y install \
  apt-transport-https \
  ca-certificates \
  curl \
  gettext \
  gnupg2 \
  jq \
  python3-pip \
  software-properties-common

RUN curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
  touch /etc/apt/sources.list.d/kubernetes.list && \
  echo "deb http://apt.kubernetes.io/ kubernetes-xenial main" | tee -a /etc/apt/sources.list.d/kubernetes.list && \
  apt-get update && \
  apt-get install --no-install-recommends -y kubectl && \
  rm -rf /var/lib/apt/lists/*

# binary will be $(go env GOPATH)/bin/golangci-lint
RUN curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(go env GOPATH)/bin v1.21.0

RUN curl -L --output /usr/local/bin/kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-linux-amd64 && chmod a+x /usr/local/bin/kind
RUN curl -L https://github.com/kyoh86/richgo/releases/download/v0.3.3/richgo_0.3.3_linux_amd64.tar.gz | tar -xz -C /tmp/ && mv /tmp/richgo /usr/local/bin

ARG kubebuilder_version=2.1.0
RUN curl -sL https://go.kubebuilder.io/dl/${kubebuilder_version}/linux/amd64 | tar -xz -C /tmp/ && mv /tmp/kubebuilder_${kubebuilder_version}_linux_amd64 /usr/local/kubebuilder

RUN go get golang.org/x/tools/cmd/goimports
RUN pip3 install pre-commit
