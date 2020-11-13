# KubeCarrier Addon for Kubermatic Kubernetes Platform

The `kubecarrier` subfolder contains KubeCarrier [addon for Kubermatic Kubernetes Platform](https://docs.kubermatic.com/kubermatic/master/advanced/addons/).
The addon contains two config options:
- `CertManager` (boolean): If enabled, cert-manager will be installed into the
target cluster as well (cert-manager is a requirement for running KubeCarrier).
- `SingleCluster` (boolean): If enabled, KubeCarrier will run in single-cluster mode, where
the same cluster  will run KubeCarrier Control Plane as well as act as a service cluster.

## Addon Maintenance
Update `kustomization.yaml` with desired image name and tag. After changing, run:
```shell script
kustomize build . > kubecarrier/operator.yaml
```

See the [KKP documentation](https://docs.kubermatic.com/kubermatic/master/advanced/addons/)
for instructions on how to enable the addon in your KKP installation.
