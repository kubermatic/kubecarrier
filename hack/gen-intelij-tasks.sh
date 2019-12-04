#!/usr/bin/env bash

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

cat << EOF > ${GIT_ROOT}/.idea/runConfigurations/kubecarrier_tender.xml
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="kubecarrier:tender" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="kubecarrier" />
    <working_directory value="\$PROJECT_DIR\$/" />
    <go_parameters value="-i -ldflags &quot;${LD_FLAGS}&quot;" />
    <parameters value="--provider-namespace=default --service-cluster-name=default --service-cluster-kubeconfig=\$USER_HOME\$/.kube/internal-kind-config-kubecarrier-svc-1" />
    <kind value="DIRECTORY" />
    <filePath value="\$PROJECT_DIR\$/|\$PROJECT_DIR\$/cmd/tender/main.go" />
    <package value="github.com/kubermatic/kubecarrier" />
    <directory value="\$PROJECT_DIR\$/cmd/tender" />
    <method v="2" />
  </configuration>
</component>
EOF
