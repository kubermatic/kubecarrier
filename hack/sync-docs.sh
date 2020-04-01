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

gh_login() {
  mkdir -p ${HOME}/.config/gh
  cat << EOF > ${HOME}/.config/gh/config.yml
github.com:
  - oauth_token: ${GITHUB_TOKEN}
EOF
}

if [[ -n ${CI:-} ]]; then
  echo "running CI setup"
  git config --global user.email "dev@loodse.com"
  git config --global user.name "Prow CI Robot"
  ensure_github_host_pubkey
fi

VERSION=${1:-$(git rev-parse --abbrev-ref HEAD)-$(git rev-parse --short HEAD)}
echo "current code version ${VERSION}"
REPO=${REPO:-git@github.com:loodse/docs.git}
if [[ -z ${WORKDIR:-} ]]; then
  WORKDIR=$(mktemp --directory)
fi
cd ${WORKDIR}

PROJECT=$(git rev-parse --show-toplevel)


if [[ ! -d ${WORKDIR}/.git ]]; then
  git clone ${REPO} ${WORKDIR}
  git checkout master
else
  echo "syncing repo"
  git checkout master
  git pull
fi

echo "================"
echo "syncing docs e.g. rsync"
echo "TODO fill this in!!!"
echo "$(date -Iseconds) ${VERSION}" > ${WORKDIR}/kubecarrier-docs-version
echo "================"

git switch --force-create ${VERSION}
git add .
# TODO fix this!
if ! git diff --cached --stat --exit-code; then
  echo "creating PR"
  git commit -a -m "updated kubecarrier docs to version ${VERSION}"
  git push --force-with-lease
  if [[ ! -f ${HOME}/.config/gh/config.yml ]]; then
    echo "gh config not properly setup; creating. The GITHUB_TOKEN env should be setup"
    gh_login
  fi
  gh pr create --fill
fi
