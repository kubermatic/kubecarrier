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

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (, $(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: \
	build-manager

# Build manager binary
build-%:
	GOOS=linux go build -o bin/$* cmd/$*/main.go

# Run the controller manager to against the kubernetes cluster.
run-manager: generate fmt vet install-manager
	go run ./cmd/manager/main.go

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./pkg/apis/...

# Install CRDs into a cluster
install-%: manifests-%
	kubectl apply -f config/$*/crd/bases

# Generate manifests e.g. CRD, RBAC etc.
manifests: \
	manifests-manager

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
	echo "running unit tests"

e2e-test:
	echo "running e2e tests"

pre-commit:
	pre-commit run -a

lint:
	golangci-lint run ./...
