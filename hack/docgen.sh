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

cat << 'EOF' > ./docs/api_reference/_index.md
---
title: API Reference
weight: 50
---

The KubeCarrier API is implemented as a extension of the [Kubernetes API](https://kubernetes.io/docs/reference/using-api/api-overview/) as `CustomResourceDefinitions`.
All available objects and their usage are described below.

The API consists of multiple API groups:
* [kubecarrier.io](#core) - Core
* [catalog.kubecarrier.io](#catalog) - Catalog
* [operator.kubecarrier.io](#operator) - Operator

EOF

# Core API Group
# --------------
echo -e "Core API"
cat << 'EOF' >> ./docs/api_reference/_index.md
## Core

The core `kubecarrier.io` API group contains the basic buildings blocks of KubeCarrier and objects to setup cross-cluster management of resources.

EOF
find ./pkg/apis/core -name '*types.go' | xargs ./bin/docgen -section-link='#core' >> ./docs/api_reference/_index.md

# Catalog API Group
# -----------------
echo -e "\nCatalog API"
cat << 'EOF' >> ./docs/api_reference/_index.md
## Catalog

The `catalog.kubecarrier.io` API group contains all objects that are used to setup service catalogs in KubeCarrier.

EOF
find ./pkg/apis/catalog -name '*types.go' | xargs ./bin/docgen -section-link='#catalog' >> ./docs/api_reference/_index.md

# Operator API Group
# ------------------
echo -e "\nOperator API"
cat << 'EOF' >> ./docs/api_reference/_index.md
## Operator

The `operator.kubecarrier.io` API group contains objects to interact with the KubeCarrier installation.

EOF
find ./pkg/apis/operator -name '*types.go' | xargs ./bin/docgen -section-link='#operator' >> ./docs/api_reference/_index.md
