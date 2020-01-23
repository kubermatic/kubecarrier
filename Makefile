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

SHELL=/bin/bash
.SHELLFLAGS=-euo pipefail -c

export CGO_ENABLED:=0
ifdef CI
	# prow sets up GOPATH really helpfully:
	# https://github.com/kubernetes/test-infra/issues/9469
	# https://github.com/kubernetes/test-infra/blob/895df89b7e4238125063157842c191dac6f7e58f/prow/pod-utils/decorate/podspec.go#L474
	export GOPATH:=${HOME}/go
	export PATH:=${PATH}:${GOPATH}/bin
endif

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}
BUILD_DATE=$(shell date +%s)
IMAGE_ORG?=quay.io/kubecarrier
MODULE=github.com/kubermatic/kubecarrier
LD_FLAGS=-X $(MODULE)/pkg/internal/version.Version=$(VERSION) -X $(MODULE)/pkg/internal/version.Branch=$(BRANCH) -X $(MODULE)/pkg/internal/version.Commit=$(SHORT_SHA) -X $(MODULE)/pkg/internal/version.BuildDate=$(BUILD_DATE)
KIND_CLUSTER?=kubecarrier
COMPONENTS = operator manager ferry catapult

all: \
	bin/linux_amd64/anchor \
	bin/darwin_amd64/anchor \
	bin/windows_amd64/anchor \
	bin/linux_amd64/operator \
	bin/linux_amd64/manager

bin/linux_amd64/%: GOARGS = GOOS=linux GOARCH=amd64
bin/darwin_amd64/%: GOARGS = GOOS=darwin GOARCH=amd64
bin/windows_amd64/%: GOARGS = GOOS=windows GOARCH=amd64

bin/%: FORCE
	$(eval COMPONENT=$(shell basename $*))
	$(GOARGS) go build -ldflags "-w $(LD_FLAGS)" -o bin/$* cmd/$(COMPONENT)/main.go

FORCE:

clean: e2e-test-clean
	rm -rf bin/$*
.PHONEY: clean

# Generate code
generate:
	@hack/codegen.sh
	# regenerate golden files to update tests
	FIX_GOLDEN=1 go test ./pkg/internal/resources/...

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

test:
	CGO_ENABLED=1 go test -race -v ./...
.PHONY: test

install:
	go install -ldflags "-w $(LD_FLAGS)" ./cmd/anchor
.PHONY: install

TEST_ID?=1
MASTER_KIND_CLUSTER?=kubecarrier-${TEST_ID}
SVC_KIND_CLUSTER?=kubecarrier-svc-${TEST_ID}

e2e-setup: install require-docker
	@unset KUBECONFIG
	@kind create cluster --name=${MASTER_KIND_CLUSTER} || true
	@kind create cluster --name=${SVC_KIND_CLUSTER} || true
	@kind get kubeconfig --internal --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${MASTER_KIND_CLUSTER}"
	@kind get kubeconfig --internal --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${SVC_KIND_CLUSTER}"
	@kind get kubeconfig --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER}"
	@kind get kubeconfig --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER}"
	@echo "kind clusters created"
	@echo "Loading the images"
	@$(MAKE) KIND_CLUSTER=${MASTER_KIND_CLUSTER} cert-manager
	@$(MAKE) KIND_CLUSTER=${MASTER_KIND_CLUSTER} kind-load

e2e-test: e2e-setup
	@go run -ldflags "-w $(LD_FLAGS)" ./cmd/anchor e2e-test run --test.v --test-id=${TEST_ID} | richgo testfilter

.PHONY: e2e-test

e2e-test-clean:
	@kind delete cluster --name=${MASTER_KIND_CLUSTER} || true
	@kind delete cluster --name=${SVC_KIND_CLUSTER} || true
.PHONY: e2e-test-clean

lint:
	pre-commit run -a
	golangci-lint run ./... --deadline=15m

tidy:
	go mod tidy

push-images: $(addprefix push-image-, $(COMPONENTS))

# build all container images except the test image
build-images: $(addprefix build-image-, $(COMPONENTS))

kind-load: $(addprefix kind-load-, $(COMPONENTS))

build-image-test: require-docker
	@mkdir -p bin/image/test
	@cp -a config/dockerfiles/test.Dockerfile bin/image/test/Dockerfile
	@cp -a .pre-commit-config.yaml bin/image/test
	@cp -a go.mod go.sum hack/start-docker.sh bin/image/test
	@docker build -t ${IMAGE_ORG}/test bin/image/test

push-image-test: build-image-test require-docker
	@docker push ${IMAGE_ORG}/test
	@echo pushed ${IMAGE_ORG}/test

.SECONDEXPANSION:
# copy binary in new folder, so docker build is only sending the binary to the docker deamon
build-image-%: bin/linux_amd64/$$* require-docker
	@mkdir -p bin/image/$*
	@mv bin/linux_amd64/$* bin/image/$*
	@cp -a config/dockerfiles/$*.Dockerfile bin/image/$*/Dockerfile
	@docker build -t ${IMAGE_ORG}/$*:${VERSION} bin/image/$*

push-image-%: build-image-$$* require-docker
	@docker push ${IMAGE_ORG}/$*:${VERSION}
	@echo pushed ${IMAGE_ORG}/$*:${VERSION}

kind-load-%: build-image-$$*
	kind load docker-image ${IMAGE_ORG}/$*:${VERSION} --name=${KIND_CLUSTER}

require-docker:
	@docker ps > /dev/null 2>&1 || start-docker.sh || (echo "cannot find running docker daemon nor can start new one" && false)
	@[[ -z "${QUAY_IO_USERNAME}" ]] || ( echo "logging in to ${QUAY_IO_USERNAME}" && docker login -u ${QUAY_IO_USERNAME} -p ${QUAY_IO_PASSWORD} quay.io )
.PHONEY: require-docker

generate-ide-tasks:
	@go run ./hack/gen-tasks.go -ldflags "${LD_FLAGS}"

install-git-hooks:
	pre-commit install
	printf "#!/bin/bash\\nmake generate-ide-tasks" > .git/hooks/post-commit && chmod +x .git/hooks/post-commit
	cp .git/hooks/post-commit .git/hooks/post-checkout
	cp .git/hooks/post-commit .git/hooks/post-merge

# Install cert-manager in the configured Kubernetes cluster
cert-manager:
	kubectl create namespace cert-manager
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.11.0/cert-manager.yaml
