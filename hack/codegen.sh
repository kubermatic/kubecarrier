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

if [ -z $(go env GOBIN) ]; then
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
# CRDs
# The `|| true` is because the controller-gen will error out if CRD_types.go embeds CustomResourcDefinition, and it will be handled in the following yq removements.
$CONTROLLER_GEN crd:crdVersions=${CRD_VERSION} paths="./pkg/apis/core/..." output:crd:artifacts:config=config/internal/manager/crd/bases || true
# The `|| true` is because the controller-gen will error out if CRD_types.go embeds CatalogEntry embeds CustomResourceValidation, and it will be handled in the following yq removements.
$CONTROLLER_GEN crd:crdVersions=${CRD_VERSION} paths="./pkg/apis/catalog/..." output:crd:artifacts:config=config/internal/manager/crd/bases || true
# Webhooks
$CONTROLLER_GEN webhook paths="./pkg/manager/internal/webhooks/..." output:webhook:artifacts:config=config/internal/manager/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/manager/..." output:rbac:artifacts:config=config/internal/manager/rbac
# Remove properties to make the CustomResourceDefinitionDiscovery yaml configuration (which embeds CustomResourceValidation) to pass the schema checks
out=$(mktemp)
yq -Y  "del(.spec.versions[].schema.openAPIV3Schema.properties.status.properties.crd.properties)" "config/internal/manager/crd/bases/kubecarrier.io_customresourcedefinitiondiscoveries.yaml" > $out && mv ${out} "config/internal/manager/crd/bases/kubecarrier.io_customresourcedefinitiondiscoveries.yaml"
yq -Y  "del(.spec.versions[].schema.openAPIV3Schema.properties.status.properties.crd.required)" "config/internal/manager/crd/bases/kubecarrier.io_customresourcedefinitiondiscoveries.yaml" > $out && mv ${out} "config/internal/manager/crd/bases/kubecarrier.io_customresourcedefinitiondiscoveries.yaml"

# Remove properties to make the CatalogEntry and Offering yaml configuration (which embeds CustomResourceValidation) to pass the schema checks
cat config/internal/manager/crd/bases/catalog.kubecarrier.io_catalogentries.yaml | yq -Y 'del(.spec.versions[].schema.openAPIV3Schema.properties.status.properties.crds.items.properties.versions.items.properties.schema.properties)' > config/internal/manager/crd/bases/catalog.kubecarrier.io_catalogentries.yaml.tmp
mv config/internal/manager/crd/bases/catalog.kubecarrier.io_catalogentries.yaml.tmp config/internal/manager/crd/bases/catalog.kubecarrier.io_catalogentries.yaml
cat config/internal/manager/crd/bases/catalog.kubecarrier.io_offerings.yaml | yq -Y 'del(.spec.versions[].schema.openAPIV3Schema.properties.offering.properties.crds.items.properties.versions.items.properties.schema.properties)' > config/internal/manager/crd/bases/catalog.kubecarrier.io_offerings.yaml.tmp
mv config/internal/manager/crd/bases/catalog.kubecarrier.io_offerings.yaml.tmp config/internal/manager/crd/bases/catalog.kubecarrier.io_offerings.yaml
statik-gen manager config/internal/manager

# Ferry
# -------
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/ferry/..." output:rbac:artifacts:config=config/internal/ferry/rbac
# The `|| true` is because the `,s/ClusterRole/Role/g` will error out if there is no match of `ClusterRole` (eg., the file is empty) in the file.
ed config/internal/ferry/rbac/role.yaml <<EOF || true
,s/ClusterRole/Role/g
w
EOF
# Statik (run only when file CONTENT has changed)
statik-gen ferry config/internal/ferry

# Catapult
# -------
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/catapult/..." output:rbac:artifacts:config=config/internal/catapult/rbac
# The `|| true` is because the `,s/ClusterRole/Role/g` will error out if there is no match of `ClusterRole` (eg., the file is empty) in the file.
ed config/internal/catapult/rbac/role.yaml <<EOF || true
,s/ClusterRole/Role/g
w
EOF
statik-gen catapult config/internal/catapult

# Elevator
# -------
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/elevator/..." output:rbac:artifacts:config=config/internal/elevator/rbac
# The `|| true` is because the `,s/ClusterRole/Role/g` will error out if there is no match of `ClusterRole` (eg., the file is empty) in the file.
ed config/internal/elevator/rbac/role.yaml <<EOF || true
,s/ClusterRole/Role/g
w
EOF
statik-gen elevator config/internal/elevator
