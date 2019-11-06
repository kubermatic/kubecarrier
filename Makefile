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

DOCKER_TEST_IMAGE?=quay.io/kubecarrier/test

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}

version:
	@echo -n ${VERSION}

test:
	go test -race -v ./...
.PHONY: test

e2e-test:
	go run ./cmd/anchor e2e-test kind-setup
	go run ./cmd/anchor e2e-test run
	go run ./cmd/anchor e2e-test kind-teardown
.PHONY: e2e-test

pre-commit:
	pre-commit run -a

lint:
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
