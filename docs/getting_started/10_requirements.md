---
title: Requirements
weight: 10
pre: <b>0. </b>
slug: requirements
date: 2020-04-24T09:00:00+02:00
---

## 0. Requirements

To install KubeCarrier you will need a Kubernetes Cluster with the [cert-manager](https://cert-manager.io/docs/) installed.

|Component   |Version       |
|------------|--------------|
|Kubernetes  | v1.16, v1.17 |
|cert-manager| v0.13.0      |

## Kubernetes Clusters

If you just want to try out KubeCarrier, we are recommending:
[kind - Kubernetes IN Docker](https://github.com/kubernetes-sigs/kind)

With kind, you can quickly spin up multiple Kubernetes Clusters for testing.

```bash
# Management Cluster
$ kind create cluster --name=kubecarrier
Creating cluster "kubecarrier" ...
 ✓ Ensuring node image (kindest/node:v1.17.0) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kubecarrier"
You can now use your cluster with:

kubectl cluster-info --context kind-kubecarrier

Have a question, bug, or feature request? Let us know! https://kind.sigs.k8s.io/#community 🙂

# kind is configuring kubectl for you:
$ kubectl config current-context
kind-kubecarrier
```

### Deploy cert-manager

``` bash
# deploy cert-manager
$ kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v0.14.0/cert-manager.yaml
# wait for it to be ready (optional)
$ kubectl wait --for=condition=available deployment/cert-manager -n cert-manager --timeout=120s
$ kubectl wait --for=condition=available deployment/cert-manager-cainjector -n cert-manager --timeout=120s
$ kubectl wait --for=condition=available deployment/cert-manager-webhook -n cert-manager --timeout=120s
```
