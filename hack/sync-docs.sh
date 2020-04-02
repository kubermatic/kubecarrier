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

ensure_github_host_pubkey() {
  # check whether we already have a known_hosts entry for Github
  if ssh-keygen -F github.com >/dev/null 2>&1; then
    echo " [*] Github's SSH host key already present" >/dev/stderr
  else
    local github_rsa_key
    # https://help.github.com/en/github/authenticating-to-github/githubs-ssh-key-fingerprints
    github_rsa_key="github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="

    echo " [*] Adding Github's SSH host key to known hosts" >/dev/stderr
    mkdir -p "$HOME/.ssh"
    chmod 700 "$HOME/.ssh"
    echo "$github_rsa_key" >> "$HOME/.ssh/known_hosts"
    chmod 600 "$HOME/.ssh/known_hosts"
  fi
}

if [[ -n ${CI:-} ]]; then
  echo "running CI setup"
  git config --global user.email "dev@loodse.com"
  git config --global user.name "Prow CI Robot"
  git config --global core.sshCommand 'ssh -o CheckHostIP=no -i /ssh/id_rsa'
  ensure_github_host_pubkey
fi

# what is the version of the code we're pushing to the ${TARGET_BRANCH}
# e.g. master, v0.1.0, etc
#
DEFAULT_VERSION=$(git rev-parse --abbrev-ref HEAD)
DEFAULT_VERSION=${DEFAULT_VERSION/release-/}
KUBECARRIER_VERSION=${KUBECARRIER_VERSION:-${DEFAULT_VERSION}}
COMMIT_SHA=$(git rev-parse HEAD)
COMMIT_MESSAGE=$(git show -s --format=%s)
REPO=${REPO:-git@github.com:loodse/docs.git}
if [[ -z ${WORKDIR:-} ]]; then
  WORKDIR=$(mktemp --directory)
fi
PROJECT=$(git rev-parse --show-toplevel)

# Which branch in the ${REPO} are we going to push are changes to?
# The diff is calculated from the master branch
TARGET_BRANCH=${TARGET_BRANCH:-master}

echo "========================"
echo "KUBECARRIER_VERSION = ${KUBECARRIER_VERSION}  # For which version of the docs to upsert"
echo "REPO                = ${REPO} # Which repo to publish the docs"
echo "WORKDIR             = ${WORKDIR}  # Where ${REPO} shall be cloned and operated on"
echo "TARGET_BRANCH       = ${TARGET_BRANCH}  # Which branch in ${REPO} to push changes to (it's branched from master)"
echo "syncing kubecarrier version ${KUBECARRIER_VERSION} to ${REPO} branch ${TARGET_BRANCH}"
echo "========================"

cd ${WORKDIR}
if [[ ! -d ${WORKDIR}/.git ]]; then
  git clone ${REPO} ${WORKDIR}
  git checkout master
else
  echo "syncing repo"
  git checkout master
  git pull
fi

mkdir -p ${WORKDIR}/content/kubecarrier/${KUBECARRIER_VERSION}
rsync -rh --delete ${PROJECT}/docs/ ${WORKDIR}/content/kubecarrier/${KUBECARRIER_VERSION}/

echo "dd"
git branch -D ${TARGET_BRANCH} || true
git checkout -b ${TARGET_BRANCH}
echo "aa"
git add .
if ! git diff --cached --stat --exit-code; then
  git commit -a -m "updated kubecarrier docs for version ${KUBECARRIER_VERSION} to commit ${COMMIT_SHA}"
  git push --force-with-lease --set-upstream origin ${TARGET_BRANCH}
fi
