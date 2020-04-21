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

# Check if the current commit has a matching tag.
COMMIT=`git rev-parse HEAD`

# git tag
TAG=$(git describe --exact-match --abbrev=0 --tags ${COMMIT} 2> /dev/null || true)

SUFFIX=''
DETAIL=''

if [ -z "$TAG" ]; then
  # Use the last tag as base
  TAG=$(git describe --abbrev=0)
  if [ -z "$TAG" ]; then
    # if tag is still empty default to v0.0.0
    # this will happen in CI/CD often
    TAG="v0.0.0"
  fi

  COMMIT_COUNT=$(git --no-pager log ${TAG}..HEAD --oneline | wc -l)
  COMMIT_COUNT_PADDED=$(printf %03d $COMMIT_COUNT)
  COMMIT_SHORT_SHA=$(git rev-parse --short HEAD)
  BRANCH=$(git rev-parse --abbrev-ref HEAD | sed 's/\//-/')

  # development version
  SUFFIX="-${BRANCH}"
  DETAIL=".${COMMIT_COUNT_PADDED}.${COMMIT_SHORT_SHA}"
else
  VERSION=$TAG
fi

# check for changed files (not untracked files)
if [ -n "$(git diff --shortstat 2> /dev/null | tail -n1)" ]; then
  # build from a dirty tree
  SUFFIX="-dirty"
fi

VERSION="${TAG}${SUFFIX}${DETAIL}"
echo -n $VERSION
