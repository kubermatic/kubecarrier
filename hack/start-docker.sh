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

# This is needed so Docker-In-Docker still works when the peer doesn't allow ICMP packages and hence path mtu discovery cant work
# Most notably, pmtud doesn't work with the hoster of the Alpine package mirror, fastly, causing dind builds of alpine to hang
# forever. Upstream issue: See https://github.com/gliderlabs/docker-alpine/issues/307#issuecomment-427246497
iptables -t mangle -A POSTROUTING -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --clamp-mss-to-pmtu

echo "Docker in Docker enabled, initializing..."
printf '=%.0s' {1..80}; echo

# If we have opted in to docker in docker, start the docker daemon,
service docker start
# the service can be started but the docker socket not ready, wait for ready
WAIT_N=0
MAX_WAIT=5
while true; do
    # docker ps -q should only work if the daemon is ready
    docker ps -q > /dev/null 2>&1 && break
    if [[ ${WAIT_N} -lt ${MAX_WAIT} ]]; then
        WAIT_N=$((WAIT_N+1))
        echo "Waiting for docker to be ready, sleeping for ${WAIT_N} seconds."
        sleep ${WAIT_N}
    else
        echo "Reached maximum attempts, not waiting any longer..."
        cat /var/log/docker.log
        exit 1
    fi
done
printf '=%.0s' {1..80}; echo
echo "Done setting up docker in docker."
