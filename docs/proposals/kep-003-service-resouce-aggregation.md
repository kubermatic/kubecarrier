# [KEP-003] Service resource aggregation

## Background

Currently, we're creating ferry deployment per each `CustromResourceDiscovery` object. The ferry component transports custom resources from the service cluster to the management cluster, and vice versa, asynchronously. For example the `db.example.io` apiGroup in the service cluster `svc-1` becomes `db.internal.svc-1.provider-1` apiGroup in the management cluster.

Furthermore, ferry component handles the webhooks for internal apiGroup. If there is `CustromResourceDiscoverySet` there's a new ferry deployment per each service cluster offering the same service.

* See also:

https://github.com/kubermatic/kubecarrier/issues/316

## Issues

The issues with this approach are multi-fold:

* propagation is asynchronously, whereas management clusters serves as a "cache" to service cluster's ground truth. In more complicated setups the management cluster's resource version and service cluster resource version could diverge in unreconciable way. To give a simple, albeit toy example consider the following:

    * There's deployment on the service cluster. This resource is propagated to the management cluster. At the start the deployment replicas is 1.
       * HPA controller on the service cluster increases the replica count to 2
       * User on the management cluster changes the deployment image

   Now we have to two diverging versions what the deployment should be, and how should it look like. Somehow we have to perform 2-way, or 3-way merge between those objects adding additional complexity.

   Note: In the common case the service cluster operators would only change the `.status` subresource, while the management cluster would handle the `.spec` field.

* webhook complexity

    We're running semi-complex code for object updating and creation webhooks. We rely on `dry-run` support from the service cluster. Service cluster resource might not support some critical `dry-run` operation, thus our webhooks won't behave as if they are being run in the service cluster in the first place.

* sysadmin operator automation
    * Where `CustomResourceDiscoverySet` is used for exposing same service from multiple service cluster, there's no unified resource on the management cluster for automation. There's `db.internal.svc-1.provider-1`, `db.internal.svc-2.provider-1`, ... api groups. The sysadmin wants something simple, `db.internal.provider-1` apiGroup smartly aggregating each service cluster resource into an unified apiGroup.

* resource consumption
    * multiple deployments per `CustomResourceDiscovery` start stacking up, increasing the KubeCarrier solution heaviness and limiting efficient resource allocation. Lean is green.


## Proposed solution

Instead of creating CRDs for each `CustomResourceDiscovery`, and deploying ferry per `CustomResourceDiscovery`, we deploy a single `neoferry` per provider. This `neoferry` is API Exstension server smartly proxying each API group to the service cluster, as well aggregating multiple service cluster under same apiGroup. The benefits are as follows:

* sync propagation --> there's no version divergence, we only provide a view in the management cluster with custom RBAC sprinkled on top.
* webhook complexity --> Under the hood original webhooks are ran on the service cluster, thus we're staying as close to the true north as possible.
* sysadmin operator automation --> we aggregate all same resource under different service cluster under common APIGroup
* resource consumption --> single deployment is less resource intensive then multiple deployments
* No adoption controllers --> This also all drawbacks listed in
https://github.com/kubermatic/kubecarrier/issues/316

The following things need to be rewritten for each proxied resource:

* `.metadata.clusterName` determines which service cluster the original resource belongs on
* `.metadata.managedFields[].apiVersion` needs to be rewritten for proper version
* `.apiVersion` needs be rewritten as well

## Implementation details

TODO
