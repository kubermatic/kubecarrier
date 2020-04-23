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
export VAULT_ADDR="https://vault.loodse.com"
if [[ ${VAULT_TOKEN:-x} == 'x' ]]; then
  echo "$$VAULT_TOKEN not set"
  exit 1
fi

echo "This script setups the master cluster pointed by KUBECONFIG=${KUBECONFIG}"

echo "export VAULT_ADDR=${VAULT_ADDR}"
export AWS_DEFAULT_REGION=eu-central-1

# shellcheck disable=SC1090
source <(vault kv get -field=aws_dns_user dev/kubecarrier)

kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.14.0/cert-manager.yaml
helm upgrade dex stable/dex  --install  --values ./dex_values.yaml

kubectl wait --for=condition=available deployment/cert-manager -n cert-manager --timeout=120s
kubectl wait --for=condition=available deployment/cert-manager-cainjector -n cert-manager --timeout=120s
kubectl wait --for=condition=available deployment/cert-manager-webhook -n cert-manager --timeout=120s

kubectl secret genetic route53-aws "--from-literal=id=${AWS_ACCESS_KEY_ID}" "--from-literal=secret=${AWS_SECRET_ACCESS_KEY}" --dry-run -o yaml | kubectl apply -f -
cat << EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1alpha3
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: neven@loodse.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - selector:
        dnsZones:
        - lab.kubecarrier.io
      dns01:
        providers:
        - name: route53
          route53:
            region: us-central-1
            accessKeyID: "${AWS_ACCESS_KEY_ID}"
            secretAccessKeySecretRef:
              name:
                name: route53-aws
                key: secret
            hostedZoneID: Z04311331EAB7S9MUWBEW
EOF

kubectl apply -f dex_ingress.yaml
