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
PROJECT=$(git rev-parse --show-toplevel)
if [[ ! -d ${PROJECT}/loodse-docs ]]; then
  git clone git@github.com:loodse/docs.git ${PROJECT}/loodse-docs
fi

cd ${PROJECT}/loodse-docs
git checkout -f kubecarrier
git pull

ln -sf ${PROJECT}/docs ${PROJECT}/loodse-docs/content/kubecarrier/master
echo "The docs shall be served at http://localhost:1313/kubecarrier/master"
make -C ${PROJECT}/loodse-docs preview
