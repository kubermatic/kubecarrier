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
