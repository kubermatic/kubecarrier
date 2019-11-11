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

export CGO_ENABLED:=0

BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
SHORT_SHA=$(shell git rev-parse --short HEAD)
VERSION?=${BRANCH}-${SHORT_SHA}
BUILD_DATE=$(shell date +%s)
IMAGE_ORG?=quay.io/kubecarrier
DOCKER_TEST_IMAGE?=quay.io/kubecarrier/test
MODULE=github.com/kubermatic/kubecarrier
LD_FLAGS="-w -X '$(MODULE)/pkg/version.Version=$(VERSION)' -X '$(MODULE)/pkg/version.Branch=$(BRANCH)' -X '$(MODULE)/pkg/version.Commit=$(SHORT_SHA)' -X '$(MODULE)/pkg/version.BuildDate=$(BUILD_DATE)'"

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (, $(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

bin/linux_amd64/%: GOARGS = GOOS=linux GOARCH=amd64
bin/darwin_amd64/%: GOARGS = GOOS=darwin GOARCH=amd64

bin/%: FORCE generate
	$(eval COMPONENT=$(shell basename $*))
	$(GOARGS) go build -ldflags $(LD_FLAGS) -o bin/$* cmd/$(COMPONENT)/main.go

FORCE:

clean:
	rm -rf bin/$*
.PHONEY: clean

# Generate code
generate: controller-gen
	go generate ./...
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/apis/...

install: \
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


push-images: \
	push-image-operator

# build all container images except the test image
build-images: \
	build-image-operator

kind-load: \
	kind-load-operator

build-image-test:
	# @echo test
	@mkdir -p bin/image/test
	@cp -a config/dockerfiles/test.Dockerfile bin/image/test/Dockerfile
	@cp -a go.mod bin/image/test
	@cp -a go.sum bin/image/test
	@docker build -t ${IMAGE_ORG}/test bin/image/test

push-image-test:
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
	kind load docker-image ${IMAGE_ORG}/$*:${VERSION} --name=kind
