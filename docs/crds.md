# Working with CustomResourceDefinitions in KubeCarrier

### CustomResourceDefinitionDiscovery

`CustomResourceDefinitionDiscovery` objects are used to fetch a CRD from a remote `ServiceCluster` and controls the registration of the remote CRD into the master cluster.

```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: CustomResourceDefinitionDiscovery
metadata:
  name: couchdbs.eu-west-1
  namespace: provider-example-cloud
spec:
  # references a CRD in the ServiceCluster
  crd:
    name: couchdbs.couchdb.io
  # references the ServiceCluster
  serviceCluster:
    name: eu-west-1

  kindOverride: CouchDBInternal

status:
  phase: Ready
  conditions:
  # is True when the CRD was found and was registered into the master cluster
  - type: Ready
    status: "True"
    reason: CRDRegistered
  # is True when the CRD can be found on the ServiceCluster,
  # is signifying the handover from the Ferry component to the main KubeCarrier controller manager
  - type: Discovered
    status: "True"
    reason: CRDFound
  observedGeneration: 10

  crd:
    apiVersion: apiextensions.k8s.io/v1beta1
    kind: CustomResourceDefinition
    metadata:
      annotations:
        controller-gen.kubebuilder.io/version: v0.2.4
      name: couchdb.couchdb.io
    spec:
      group: couchdb.io
      names:
        kind: CouchDB
        listKind: CouchDBList
        plural: couchdbs
        singular: couchdb
      # ...
    status: {}
```

The `CustomResourceDefinitionDiscovery` object above will register this CRD:

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    # Provider name
    kubecarrier.io/provider: example.cloud
    # ServiceCluster name
    kubecarrier.io/service-cluster: eu-west-1

  # name is singular kind + "." + group
  name: couchdbinternal.eu-west-1.example.cloud
spec:
  # group is ServiceCluster Name + "." + Provider Name
  group: eu-west-1.example.cloud
  names:
    # uses CRDDiscovery.spec.kindOverride and generates listKind/plural/singular as in:
    # https://github.com/kubernetes-sigs/controller-tools/blob/v0.2.4/pkg/crd/spec.go#L58
    kind: CouchDBInternal
    listKind: CouchDBInternalList
    plural: couchdbinternals
    singular: couchdbinternal
status: {}
```

### DerivedCustomResourceDefinition

`DerivedCustomResourceDefinition` is deriving a new CRD from a given one.

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: DerivedCustomResourceDefinition
metadata:
  name: couchdbs.eu-west-1
  namespace: provider-example-cloud
spec:
  crd:
    name: couchdbinternal.eu-west-1.example.cloud

  kindOverride: CouchDB

  expose:
  - versions:
    # versions of the CRD that this expose config can be applied to
    - v1alpha1
    # fields that are exposed on the external/public instance
    fields:
    - jsonPath: .spec.username
    - jsonPath: .spec.password
    - jsonPath: .status.address
    - jsonPath: .status.fauxtonAddress
    - jsonPath: .status.observedGeneration
    - jsonPath: .status.phase
status:
  phase: Ready
  conditions: []
  observedGeneration: 10
```

The derived CRD:

```yaml
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  labels:
    # Provider name
    kubecarrier.io/provider: example.cloud
    # ServiceCluster name
    kubecarrier.io/service-cluster: eu-west-1

  # name is singular kind + "." + group
  name: couchdb.eu-west-1.example.cloud
spec:
  # group is ServiceCluster Name + "." + Provider Name
  group: eu-west-1.example.cloud
  names:
    # uses CRDDiscovery.spec.kindOverride and generates listKind/plural/singular as in:
    # https://github.com/kubernetes-sigs/controller-tools/blob/v0.2.4/pkg/crd/spec.go#L58
    kind: CouchDB
    listKind: CouchDBList
    plural: couchdbs
    singular: couchdb
status: {}
```

### Catapult

The `Catapult` component is catapulting the dynamic CRD's from the `CRDDiscovery` and `DerivedCRD` objects from namespace to namespace and from cluster to cluster.

```yaml
apiVersion: operator.kubecarrier.io/v1alpha1
kind: Catapult
metadata:
  name: couchdbs.eu-west-1
  namespace: provider-example-cloud
spec:
  internalCRD:
    # from CRDDiscovery status
    # generated from ServiceCluster
    kind: CouchDBInternal
    version: v1alpha1
    group: eu-west-1.example.cloud

  serviceClusterCRD:
    # from CRDDiscovery status
    kind: CouchDB
    version: v1alpha1
    group: couchdb.couchdb.io

  externalCRD: # optional
    # from CRDConfig status
    kind: CouchDB
    version: v1alpha1
    group: eu-west-1.example.cloud

status:
  phase: Ready
  conditions: []
  observedGeneration: 10
```

### CustomResourceDefinitionSet

Because configuring `CRDDiscovery`, `DerivedCRD` and `Catapult` instances for multiple `ServiceClusters` is getting very boring and cumbersome, the `CustomResourceDefinitionSet` is abstracting that away and is creating all needed objects for multiple `ServiceClusters`.

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: CustomResourceDefinitionSet
metadata:
  name: couchdbs
spec:
  # selects the service clusters to work on
  serviceClusterSelector: {}

  # selects the CRD instance on each ServiceCluster
  crd:
    name: couchdbs.couchdb.io

  expose:
  - versions:
    # versions of the CRD that this expose config can be applied to
    - v1alpha1
    # fields that are exposed on the external/public instance
    fields:
    - jsonPath: .spec.username
    - jsonPath: .spec.password
    - jsonPath: .status.address
    - jsonPath: .status.fauxtonAddress
    - jsonPath: .status.observedGeneration
    - jsonPath: .status.phase
status:
  phase: Ready
  conditions: []
  observedGeneration: 10
```
