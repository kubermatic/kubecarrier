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

JOB_LOG=${PULL_NUMBER:-}-${JOB_NAME:-}-${BUILD_ID:-}
workdir=$(mktemp -d)
if [[ "${JOB_LOG}" != "--" ]]; then
  workdir=${workdir}/${JOB_LOG}
  mkdir -p ${workdir}
fi

function cleanup() {
  cat ${workdir}/test.out | go tool test2json | tee ${workdir}/test.json | go run ./hack/testjsonformat | tee ${workdir}/test.inorder.out
  echo "starting cleanup & log upload"
  mkdir -p ${workdir}/management
  mkdir -p ${workdir}/svc
  kind export logs --name ${MANAGEMENT_KIND_CLUSTER} ${workdir}/management
  kind export logs --name ${SVC_KIND_CLUSTER} ${workdir}/svc
  docker cp ${MANAGEMENT_KIND_CLUSTER}-control-plane:/var/log/kube-apiserver-audit.log ${workdir}/management/audit.log
  docker cp ${SVC_KIND_CLUSTER}-control-plane:/var/log/kube-apiserver-audit.log ${workdir}/svc/audit.log
  echo "find all logs in ${workdir}"

  # https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#job-environment-variables
  if [[ "${JOB_LOG}" != "--" ]]; then
    tmpdir=$(mktemp -d)
    zip --junk-paths --quiet -r "${tmpdir}/${JOB_LOG}.zip" "${workdir}"
    aws s3 cp "${tmpdir}/${JOB_LOG}.zip" "s3://e2elogs.kubecarrier.io/${JOB_LOG}.zip"
    echo "https://s3.eu-central-1.amazonaws.com/e2elogs.kubecarrier.io/${JOB_LOG}.zip"
  fi
}

trap cleanup EXIT
kubectl kubecarrier e2e-test run --test.timeout=10m --test.v --test.failfast --test-id=${TEST_ID} | tee ${workdir}/test.out
