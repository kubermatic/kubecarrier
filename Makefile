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
COMPONENTS = operator manager ferry catapult elevator

# https://github.com/kubernetes-sigs/kind/releases/tag/v0.7.0
# There are different node image types for newer kind v0.7.0+ version
KIND_NODE_IMAGE?=kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62
# KIND_NODE_IMAGE=kindest/node:v1.16.4@sha256:b91a2c2317a000f3a783489dfb755064177dbc3a0b2f4147d50f04825d016f55
# KIND_NODE_IMAGE=kindest/node:v1.15.7@sha256:e2df133f80ef633c53c0200114fce2ed5e1f6947477dbc83261a6a921169488d
# KIND_NODE_IMAGE=kindest/node:v1.14.10@sha256:81ae5a3237c779efc4dda43cc81c696f88a194abcc4f8fa34f86cf674aa14977

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

bin/docgen: hack/docgen/main.go
	$(GOARGS) go build -ldflags "-w $(LD_FLAGS)" -o bin/docgen ./hack/docgen

clean: e2e-test-clean
	rm -rf bin/$*
.PHONEY: clean

# Generate code
generate: docs
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
	@kind create cluster --name=${MASTER_KIND_CLUSTER} --image=${KIND_NODE_IMAGE} || true
	@kind get kubeconfig --internal --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${MASTER_KIND_CLUSTER}"
	@kind get kubeconfig --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER}"
	@echo "Deploy cert-manger in master cluster"
	# Deploy cert-manager right after the creation of the master cluster, since the deployments of cert-manger take some time to get ready.
	@$(MAKE) KUBECONFIG=${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER} cert-manager
	@kind create cluster --name=${SVC_KIND_CLUSTER} --image=${KIND_NODE_IMAGE} || true
	@kind get kubeconfig --internal --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${SVC_KIND_CLUSTER}"
	@kind get kubeconfig --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER}"
	@echo "kind clusters created"
	@kubectl --kubeconfig=${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER} apply -n default -f ./config/serviceCluster
	@kubectl create serviceaccount kubecarrier -n default --dry-run -o yaml | kubectl apply --kubeconfig=${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER} -f -
	@kubectl create clusterrolebinding kubecarrier --serviceaccount=default:kubecarrier --clusterrole kubecarrier:service-cluster-admin --dry-run -o yaml |  kubectl apply --kubeconfig=${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER} -f -
	@go run ./hack/impersonate --kubeconfig "${HOME}/.kube/internal-kind-config-${SVC_KIND_CLUSTER}" --as "system:serviceaccount:default:kubecarrier"
	@echo "service cluster service account created"
	@echo "Loading the images"
	@$(MAKE) KIND_CLUSTER=${MASTER_KIND_CLUSTER} kind-load -j 5

# soft-reinstall reinstall kubecarrier in the e2e cluster. It's intended for usage during development
soft-reinstall: e2e-setup install
	@anchor setup --kubeconfig "${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER}"
	@kubectl --kubeconfig "${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER}" delete pod --all -n kubecarrier-system

e2e-test: e2e-setup
	@LD_FLAGS="$(LD_FLAGS)" TEST_ID=${TEST_ID} MASTER_KIND_CLUSTER=${MASTER_KIND_CLUSTER} SVC_KIND_CLUSTER=${SVC_KIND_CLUSTER} $(SHELL) ./hack/.e2e-test.sh

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
	kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager.yaml
	kubectl wait --for=condition=available deployment/cert-manager -n cert-manager --timeout=120s
	kubectl wait --for=condition=available deployment/cert-manager-cainjector -n cert-manager --timeout=120s
	kubectl wait --for=condition=available deployment/cert-manager-webhook -n cert-manager --timeout=120s

docs: bin/docgen
	@find ./pkg/apis -name '*types.go' | xargs ./bin/docgen > ./docs/api_reference.md

.PHONEY: docs
