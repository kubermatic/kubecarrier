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

export CGO_ENABLED:=0

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}
BUILD_DATE=$(shell date +%s)
IMAGE_ORG?=quay.io/kubecarrier
MODULE=github.com/kubermatic/kubecarrier
LD_FLAGS="-w -X '$(MODULE)/pkg/internal/version.Version=$(VERSION)' -X '$(MODULE)/pkg/internal/version.Branch=$(BRANCH)' -X '$(MODULE)/pkg/internal/version.Commit=$(SHORT_SHA)' -X '$(MODULE)/pkg/internal/version.BuildDate=$(BUILD_DATE)'"
KIND_CLUSTER?=kubecarrier

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (, $(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: \
	bin/linux_amd64/anchor \
	bin/darwin_amd64/anchor \
	bin/windows_amd64/anchor \
	bin/linux_amd64/operator

bin/linux_amd64/%: GOARGS = GOOS=linux GOARCH=amd64
bin/darwin_amd64/%: GOARGS = GOOS=darwin GOARCH=amd64
bin/windows_amd64/%: GOARGS = GOOS=windows GOARCH=amd64

bin/%: FORCE
	$(eval COMPONENT=$(shell basename $*))
	$(GOARGS) go build -ldflags $(LD_FLAGS) -o bin/$* cmd/$(COMPONENT)/main.go

FORCE:

clean: e2e-test-clean
	rm -rf bin/$*
.PHONEY: clean

# Generate code
generate: controller-gen
	go generate ./...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate/boilerplate.go.txt,year=$(shell date +%Y) paths=./pkg/apis/...

install:
	go install -ldflags $(LD_FLAGS) ./cmd/anchor

install-crds: \
	install-operator

# Install CRDs into a cluster
install-%: manifests-%
	kubectl apply -f config/$*/crd/bases

# Generate manifests e.g. CRD, RBAC etc.
manifests: \
	manifests-operator

manifests-%: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager webhook paths="./..." output:crd:artifacts:config=config/$*/crd/bases output:rbac:artifacts:config=config/$*/rbac output:webhook:artifacts:config=config/$*/webhook

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.2
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

test:
	CGO_ENABLED=1 go test -race -v ./...
.PHONY: test

TEST_ID?=1
MASTER_KIND_CLUSTER?=kubecarrier-${TEST_ID}
SVC_KIND_CLUSTER?=kubecarrier-svc-${TEST_ID}

e2e-test: install
	@docker ps > /dev/null 2>&1 || start-docker.sh || (echo "cannot find running docker daemon nor can start new one" && false)
	@unset KUBECONFIG
	@kind create cluster --name=${MASTER_KIND_CLUSTER} || true
	@kind create cluster --name=${SVC_KIND_CLUSTER} || true
	@kind get kubeconfig --internal --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${MASTER_KIND_CLUSTER}"
	@kind get kubeconfig --internal --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/internal-kind-config-${SVC_KIND_CLUSTER}"
	@kind get kubeconfig --name=${MASTER_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${MASTER_KIND_CLUSTER}"
	@kind get kubeconfig --name=${SVC_KIND_CLUSTER} > "${HOME}/.kube/kind-config-${SVC_KIND_CLUSTER}"
	@echo "kind clusters created"
	@echo "Loading the images"
	@$(MAKE) KIND_CLUSTER=${MASTER_KIND_CLUSTER} kind-load
	@go run -ldflags $(LD_FLAGS) ./cmd/anchor e2e-test run --test.v --test-id=${TEST_ID} | richgo testfilter
.PHONY: e2e-test

e2e-test-clean:
	@kind delete cluster --name=${MASTER_KIND_CLUSTER} || true
	@kind delete cluster --name=${SVC_KIND_CLUSTER} || true
.PHONY: e2e-test-clean

lint:
	pre-commit run -a
	golangci-lint run ./...

tidy:
	go mod tidy

push-images: \
	push-image-operator

# build all container images except the test image
build-images: \
	build-image-operator

kind-load: \
	kind-load-operator

build-image-test:
	@mkdir -p bin/image/test
	@cp -a config/dockerfiles/test.Dockerfile bin/image/test/Dockerfile
	@cp -a .pre-commit-config.yaml bin/image/test
	@cp -a go.mod go.sum hack/start-docker.sh bin/image/test
	@docker build -t ${IMAGE_ORG}/test bin/image/test

push-image-test: build-image-test
	@env
	@[[ -z $${CI:-} ]] || ( start-docker.sh && docker login -u ${QUAY_IO_USERNAME} -p ${QUAY_IO_PASSWORD} quay.io )
	@docker push ${IMAGE_ORG}/test
	@echo pushed ${IMAGE_ORG}/test

.SECONDEXPANSION:
# copy binary in new folder, so docker build is only sending the binary to the docker deamon
build-image-%: bin/linux_amd64/$$*
	@mkdir -p bin/image/$*
	@mv bin/linux_amd64/$* bin/image/$*
	@cp -a config/dockerfiles/$*.Dockerfile bin/image/$*/Dockerfile
	@docker build -t ${IMAGE_ORG}/$*:${VERSION} bin/image/$*

push-image-%: build-image-$$*
	@docker push ${IMAGE_ORG}/$*:${VERSION}
	@echo pushed ${IMAGE_ORG}/$*:${VERSION}

kind-load-%: build-image-$$*
	kind load docker-image ${IMAGE_ORG}/$*:${VERSION} --name=${KIND_CLUSTER}
