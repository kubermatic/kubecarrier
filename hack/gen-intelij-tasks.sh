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
LD_FLAGS=${1//-w/}
GIT_ROOT=$(git rev-parse --show-toplevel)


cat << EOF > ${GIT_ROOT}/.idea/runConfigurations/kubecarrier_operator.xml
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="kubecarrier:operator" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="kubecarrier" />
    <working_directory value="\$PROJECT_DIR\$/" />
    <go_parameters value="-i -ldflags &quot;${LD_FLAGS}&quot;" />
    <kind value="DIRECTORY" />
    <filePath value="\$PROJECT_DIR\$/|\$PROJECT_DIR\$/cmd/operator/main.go" />
    <package value="github.com/kubermatic/kubecarrier" />
    <directory value="\$PROJECT_DIR\$/cmd/operator" />
    <method v="2" />
  </configuration>
</component>
EOF

cat << EOF > ${GIT_ROOT}/.idea/runConfigurations/kubecarrier_ferry.xml
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="kubecarrier:ferry" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="kubecarrier" />
    <working_directory value="\$PROJECT_DIR\$/" />
    <go_parameters value="-i -ldflags &quot;${LD_FLAGS}&quot;" />
    <parameters value="--provider-namespace=default --service-cluster-name=default --service-cluster-kubeconfig=\$USER_HOME\$/.kube/internal-kind-config-kubecarrier-svc-1" />
    <kind value="DIRECTORY" />
    <filePath value="\$PROJECT_DIR\$/|\$PROJECT_DIR\$/cmd/ferry/main.go" />
    <package value="github.com/kubermatic/kubecarrier" />
    <directory value="\$PROJECT_DIR\$/cmd/ferry" />
    <method v="2" />
  </configuration>
</component>
EOF
