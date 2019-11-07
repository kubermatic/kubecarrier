# Copyright 2019 The Kubecarrier Authors.
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

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}
BUILD_DATE=$(shell date +%s)
DOCKER_TEST_IMAGE?=quay.io/kubecarrier/test
MODULE=github.com/kubermatic/kubecarrier
LD_FLAGS="-w -X '$(MODULE)/pkg/version.Version=$(VERSION)' -X '$(MODULE)/pkg/version.Branch=$(BRANCH)' -X '$(MODULE)/pkg/version.Commit=$(SHORT_SHA)' -X '$(MODULE)/pkg/version.BuildDate=$(BUILD_DATE)'"

all: \
	build-anchor

# Build binaries for components
build-%:
	go build -ldflags $(LD_FLAGS) -o bin/$* cmd/$*/main.go

test:
	go test ./...

e2e-test:
	echo "running e2e tests"

lint:
	pre-commit run -a
	golangci-lint run ./...

tidy:
	go mod tidy

build-test-docker-image:
	@docker build -f ./config/dockerfiles/test.Dockerfile -t ${DOCKER_TEST_IMAGE} ./
	@echo built ${DOCKER_TEST_IMAGE}
.PHONEY: build-test-docker-image

push-test-docker-image: build-test-docker-image
	@docker push ${DOCKER_TEST_IMAGE}
	@echo pushed ${DOCKER_TEST_IMAGE}
.PHONEY: push-test-docker-image
