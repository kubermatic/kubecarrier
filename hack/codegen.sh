#!/usr/bin/env bash

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

set -euo pipefail

if [ -z $(go env GOBIN)]; then
GOBIN=$(go env GOPATH)/bin
else
GOBIN=$(go env GOBIN)
fi

if [ -z $(which controller-gen) ]; then
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.2
  CONTROLLER_GEN=$GOBIN/controller-gen
else
  CONTROLLER_GEN=$(which controller-gen)
fi

# DeepCopy functions
$CONTROLLER_GEN object:headerFile=./hack/boilerplate/boilerplate.go.txt,year=$(date +%Y) paths=./pkg/apis/...

function statik-gen {
  set -x
  local component=$1
  local src=$2
  if [ -z "$(git status --porcelain ${src})" ]; then
    echo $1: statik up-to-date
  else
    statik -src=${src} -p $1 -dest pkg/internal/resources -f -c ''
    cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - pkg/internal/resources/$1/statik.go > pkg/internal/resources/$1/statik.go.tmp
    mv pkg/internal/resources/$1/statik.go.tmp pkg/internal/resources/$1/statik.go
    echo $1: statik regenerated
  fi
}

# Operator
# --------
# CRDs/Webhooks
$CONTROLLER_GEN crd webhook paths="./pkg/apis/operator/..." output:crd:artifacts:config=config/operator/crd/bases output:webhook:artifacts:config=config/operator/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/operator/..." output:rbac:artifacts:config=config/operator/rbac
# Statik (run only when file CONTENT has changed)
statik-gen operator config/operator

# Manager
# -------
# CRDs/Webhooks
$CONTROLLER_GEN crd webhook paths="./pkg/apis/core/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
$CONTROLLER_GEN crd webhook paths="./pkg/apis/catalog/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/manager/..." output:rbac:artifacts:config=config/internal/manager/rbac
# Statik (run only when file CONTENT has changed)
statik-gen manager config/internal/manager

# Tender
# -------
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/tender/..." output:rbac:artifacts:config=config/internal/tender/rbac
# Statik (run only when file CONTENT has changed)
statik-gen tender config/internal/tender
