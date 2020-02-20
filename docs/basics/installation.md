# Installation

## Requirements

|Component   |Version       |
|------------|--------------|
|Kubernetes  | v1.16, v1.17 |
|cert-manager| v0.13.0      |

## Get KubeCarrier

A KubeCarrier installation is managed by a kubectl plugin. To get the plugin, you can either:

- **Use Krew**
  [Krew](https://github.com/kubernetes-sigs/krew/) is a package manager for kubectl plugins.
  `kubectl krew install kubecarrier`

- **Manual**
  Just visit the KubeCarrier [release page](https://github.com/kubermatic/kubecarrier/releases),
  download the archive and put the contained `kubectl-kubecarrier` binary into your `$PATH`.

To check whether the installation is working check the output of `kubectl plugin list` or run `kubectl kubecarrier version`.

## Install

To install KubeCarrier into a cluster:

#### 1. Make sure you are connected to the right cluster!

`kubectl config get-contexts` and
`kubectl config current-context` are your friend.

#### 2. Install cert-manager

The cert-manager is used to generate internal certificates to register webhooks into the `kube-apiserver`.

```
# Setup cert-manager
kubectl create namespace cert-manager
kubectl apply \
  -f https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager.yaml

# We want to wait until all components are online
kubectl wait --for=condition=available \
  -n cert-manager \
  deployment/cert-manager
kubectl wait --for=condition=available \
  -n cert-manager \
  deployment/cert-manager-cainjector
kubectl wait --for=condition=available \
  -n cert-manager \
  deployment/cert-manager-webhook
```

#### 3. Install KubeCarrier

The KubeCarrier installation is managed with our own Kubernetes Operator and the kubectl plugin is just installing this operator.

Just run `kubectl kubecarrier setup` and you are ready to go!
