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

# This file should ONLY be called from within Makefile!!!
set -euo pipefail

function cleanup() {
  local workdir=$(mktemp -d)
  mkdir -p ${workdir}/master
  mkdir -p ${workdir}/svc
  kind export logs --name ${MASTER_KIND_CLUSTER} ${workdir}/master
  kind export logs --name ${SVC_KIND_CLUSTER} ${workdir}/svc

  # https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#job-environment-variables
  local JOB_LOG=${JOB_NAME:-}-${BUILD_ID:-}
  if [[ "${JOB_LOG}" != "-" ]]; then
    zip -r "${workdir}/${JOB_LOG}.zip" "${workdir}/master" "${workdir}/svc"
    aws s3 cp "${workdir}/${JOB_LOG}.zip" "s3://e2elogs.kubecarrier.io/${JOB_LOG}.zip"
  fi
}

trap cleanup EXIT
go run -ldflags "-w ${LD_FLAGS}" ./cmd/anchor e2e-test run --test.v --test-id=${TEST_ID} | richgo testfilter
