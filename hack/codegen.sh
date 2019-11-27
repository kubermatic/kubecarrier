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

# Operator
# --------
# CRDs/Webhooks
$CONTROLLER_GEN crd webhook paths="./pkg/apis/operator/..." output:crd:artifacts:config=config/operator/crd/bases output:webhook:artifacts:config=config/operator/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/operator/..." output:rbac:artifacts:config=config/operator/rbac
# Statik (run only when file CONTENT has changed)
if [ -z "$(git status --porcelain config/operator)" ] && [[ -z ${FORCE_STATIK-} ]]; then
  echo operator: statik up-to-date
else \
  statik -src=config/operator -p operator -dest pkg/internal/resources -f -c ''
  cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - pkg/internal/resources/operator/statik.go > pkg/internal/resources/operator/statik.go.tmp
  mv pkg/internal/resources/operator/statik.go.tmp pkg/internal/resources/operator/statik.go
  echo operator: statik regenerated
fi

# Manager
# -------
# CRDs/Webhooks
$CONTROLLER_GEN crd webhook paths="./pkg/apis/core/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
$CONTROLLER_GEN crd webhook paths="./pkg/apis/catalog/..." output:crd:artifacts:config=config/internal/manager/crd/bases output:webhook:artifacts:config=config/internal/manager/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/manager/..." output:rbac:artifacts:config=config/internal/manager/rbac
# Statik (run only when file CONTENT has changed)
if [ -z "$(git status --porcelain config/internal/manager)" ] && [[ -z ${FORCE_STATIK-} ]]; then
  echo manager: statik up-to-date
else \
  statik -src=config/internal/manager -p manager -dest pkg/internal/resources -f -c ''
  cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - pkg/internal/resources/manager/statik.go > pkg/internal/resources/manager/statik.go.tmp
  mv pkg/internal/resources/manager/statik.go.tmp pkg/internal/resources/manager/statik.go
  echo manager: statik regenerated
fi

# E2E
# -------
# CRDs/Webhooks
$CONTROLLER_GEN crd webhook paths="./pkg/apis/e2e/..." output:crd:artifacts:config=config/internal/e2eoperator/crd/bases output:webhook:artifacts:config=config/internal/e2eoperator/webhook
# RBAC
$CONTROLLER_GEN rbac:roleName=manager-role paths="./pkg/e2eoperator/..." output:rbac:artifacts:config=config/internal/e2eoperator/rbac
# Statik (run only when file CONTENT has changed)
if [ -z "$(git status --porcelain config/internal/e2eoperator)" ] && [[ -z ${FORCE_STATIK-} ]]; then
  echo e2eoperator: statik up-to-date
else \
  statik -src=config/internal/e2eoperator -p e2eoperator -dest pkg/internal/resources -f -c ''
  cat hack/boilerplate/boilerplate.generatego.txt | sed s/YEAR/$(date +%Y)/ | cat - pkg/internal/resources/e2eoperator/statik.go > pkg/internal/resources/e2eoperator/statik.go.tmp
  mv pkg/internal/resources/e2eoperator/statik.go.tmp pkg/internal/resources/e2eoperator/statik.go
  echo e2eoperator: statik regenerated
fi
