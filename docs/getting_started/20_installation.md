---
title: Installation
weight: 20
pre: <b>1. </b>
slug: installation
date: 2020-04-24T09:00:00+02:00
---

KubeCarrier is distributed via a public container registry [quay.io/kubecarrier](https://quay.io/kubecarrier). While the KubeCarrier installation is managed by the KubeCarrier operator, installing and upgrading the operator is done via our kubectl plugin.

This CLI tool will gain more utility functions as the project matures.

## Install the kubectl plugin

To install the kubectl plugin, just visit the KubeCarrier [release page](https://github.com/kubermatic/kubecarrier/releases), download the archive and put the contained `kubecarrier` binary into your `$PATH` as `kubectl-kubecarrier`.

Make sure the binary is executable.

If everything worked, you should now be setup with the `kubecarrier` plugin:
*(Your version should be way newer than this example)*

```bash
$ kubectl kubecarrier version --full
branch: master
buildTime: "2020-02-25T14:03:31Z"
commit: a23bdbe
goVersion: go1.13
platform: linux/amd64
version: master-a23bdbe
```

## Install KubeCarrier

```bash
# make sure you are connected to the cluster,
# that you want to install KubeCarrier on
$ kubectl config current-context
kind-kubecarrier

# install KubeCarrier
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
0.19s ✔  Deploy KubeCarrier Operator
6.29s ✔  Deploy KubeCarrier
```

The `kubectl kubecarrier setup` command is idempotent, so its safe to just re-run it multiple times, if you encounter any error in your setup.

## Debugging

KubeCarrier is installed into the `kubecarrier-system` Namespace by default.

If a step in the installation is timing out, you should check the logs of the respective component:

### Operator
```bash
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
10.09s ✖  Deploy KubeCarrier Operator
Error: deploying kubecarrier operator: timed out waiting for the condition

$ kubectl get po -n kubecarrier-system
NAME                                          READY   STATUS   RESTARTS   AGE
kubecarrier-operator-manager-7d4b8f74-mgbgn   0/1     Error    2          32s

$ kubectl logs -n kubecarrier-system kubecarrier-operator-manager-7d4b8f74-mgbgn
[...]
Error: running manager: no matches for kind "Issuer" in version "cert-manager.io/v1alpha2"
[...]
```

In this case the cert-manager was not installed beforehand.

### KubeCarrier Control Plane
```bash
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
0.19s ✔  Deploy KubeCarrier Operator
60.09s ✖  Deploy KubeCarrier
Error: deploying kubecarrier: timed out waiting for the condition

$ kubectl get po -n kubecarrier-system
NAME                                                      READY   STATUS             RESTARTS   AGE
kubecarrier-manager-controller-manager-56bfd4dcbd-8rg4l   1/1     CrashLoopBackOff   0          11m
kubecarrier-operator-manager-7d4b8f74-vfsxl               1/1     Running            0          11m

$ kubectl logs -n kubecarrier-system kubecarrier-manager-controller-manager-56bfd4dcbd-8rg4l
```
