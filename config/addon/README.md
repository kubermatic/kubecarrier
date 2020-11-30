# KubeCarrier Addon for Kubermatic Kubernetes Platform

The `kubecarrier` subfolder contains the KubeCarrier
[addon for Kubermatic Kubernetes Platform](https://docs.kubermatic.com/kubermatic/master/advanced/addons/).

The addon is dependent on cert-manger, which can be installed as a separate addon located in the `cert-manager` subfolder.

The KubeCarrier addon provides one config option:
- `SingleCluster` (boolean): If enabled, KubeCarrier will run in single-cluster mode, where
the same cluster  will run KubeCarrier Control Plane as well as act as a service cluster.

## Example Addon Resources
For the cert-manger addon:
```yaml
apiVersion: kubermatic.k8s.io/v1
kind: Addon
metadata:
  name: cert-manager
  namespace: cluster-bjg9qmhctj
spec:
  cluster:
    kind: Cluster
    name: bjg9qmhctj
    uid: 30c13f8b-235e-4512-8673-d9b0f3a41f27
  name: cert-manager
```

For the KubeCarrier addon (with `SingleCluster` mode enabled and a dependency on cert-manager CRD):
```yaml
apiVersion: kubermatic.k8s.io/v1
kind: Addon
metadata:
  name: kubecarrier
  namespace: cluster-bjg9qmhctj
spec:
  cluster:
    kind: Cluster
    name: bjg9qmhctj
    uid: 30c13f8b-235e-4512-8673-d9b0f3a41f27
  name: kubecarrier
  variables: {"SingleCluster":true}
  requiredResourceTypes:
    - Group: cert-manager.io
      Version: v1alpha2
      Kind: CertificateRequest
```

## Addon Maintenance
Update `kustomization.yaml` with desired image name and tag. After changing, run:
```shell script
kustomize build . > kubecarrier/operator.yaml
```

See the [KKP documentation](https://docs.kubermatic.com/kubermatic/master/advanced/addons/)
for instructions on how to enable the addon in your KKP installation.
