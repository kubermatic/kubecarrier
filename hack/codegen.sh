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
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4
  CONTROLLER_GEN=$GOBIN/controller-gen
else
  CONTROLLER_GEN=$(which controller-gen)
fi

CONTROLLER_GEN_VERSION=$(${CONTROLLER_GEN} --version)
CONTROLLER_GEN_WANT_VERSION="Version: v0.2.4"

if [[  ${CONTROLLER_GEN_VERSION} != ${CONTROLLER_GEN_WANT_VERSION} ]]; then
  echo "Wrong controller-gen version. Wants ${CONTROLLER_GEN_WANT_VERSION} found ${CONTROLLER_GEN_VERSION}"
  exit 1
fi

function statik-gen {
  local component=$1
  local src=$2
  if [ -z "$(git status --porcelain ${src})" ] && [[ -z ${FORCE_STATIK:-} ]]; then
    echo ${component}: statik up-to-date
  else
    statik -src=${src} -p ${component} -dest pkg/internal/resources -f -c ''
    cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - pkg/internal/resources/${component}/statik.go > pkg/internal/resources/${component}/statik.go.tmp
    mv pkg/internal/resources/${component}/statik.go.tmp pkg/internal/resources/${component}/statik.go
    echo ${component}: statik regenerated
  fi
}

# DeepCopy functions
$CONTROLLER_GEN object:headerFile=./hack/boilerplate/boilerplate.go.txt,year=$(date +%Y) paths=./pkg/apis/...

CRD_VERSION="v1"

# Operator
# --------
# CRDs/Webhooks
$CONTROLLER_GEN crd:crdVersions=${CRD_VERSION} webhook paths="./pkg/apis/operator/..." output:crd:artifacts:config=config/operator/crd/bases output:webhook:artifacts:config=config/operator/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/operator/..." output:rbac:artifacts:config=config/operator/rbac
statik-gen operator config/operator

# Manager
# -------
# CRDs/Webhooks
$CONTROLLER_GEN crd:crdVersions=${CRD_VERSION} webhook paths="./pkg/apis/core/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
$CONTROLLER_GEN crd:crdVersions=${CRD_VERSION} webhook paths="./pkg/apis/catalog/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/manager/..." output:rbac:artifacts:config=config/internal/manager/rbac
statik-gen manager config/internal/manager

# Tender
# -------
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/tender/..." output:rbac:artifacts:config=config/internal/tender/rbac
sed -i 's/ClusterRole/Role/g' config/internal/tender/rbac/role.yaml
# Statik (run only when file CONTENT has changed)
statik-gen tender config/internal/tender
