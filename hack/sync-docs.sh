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


VERSION=${1:-$(git rev-parse --abbrev-ref HEAD)-$(git rev-parse --short HEAD)}
echo "current code version ${VERSION}"
REPO=${REPO:-git@github.com:loodse/docs.git}
if [[ -z ${WORKDIR:-} ]]; then
  WORKDIR=$(mktemp --directory)
fi

PROJECT=$(git rev-parse --show-toplevel)

if [[ ! -d ${WORKDIR}/.git ]]; then
  git clone ${REPO} ${WORKDIR}
else
  echo "syncing repo"
  pushd ${WORKDIR}
  git checkout master
  # git pull
  popd
fi

pushd ${WORKDIR}

echo "syncing docs e.g. rsync"
echo "$(date -Iseconds) ${VERSION}" > ${WORKDIR}/kubecarrier-docs-version

if [[ ! -f ${HOME}/.config/gh/config.yml ]]; then
  echo "gh config not properly setup; creating. The GITHUB_TOKEN env should be setup"
  mkdir -p ${HOME}/.config/gh
cat << EOF > ${HOME}/.config/gh/config.yml
github.com:
  - oauth_token: ${GITHUB_TOKEN}
EOF
fi

pushd ${WORKDIR}
git switch --force-create ${VERSION}
git add .
git commit -a -m "updated kubecarrier docs to version ${VERSION}"
git push --force-with-lease
gh pr create --fill
popd
