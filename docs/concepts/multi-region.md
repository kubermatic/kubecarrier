# KubeCarrier - Multi Region API

## Concepts

### Regions

KubeCarrier treats every `ServiceCluster` as a Region. Each region can have differences in object schemas and are exposed to Tenants. They are similar in scope to regions in AWS or GCP.

### RegionClaims

`RegionClaim` objects claim a region for a KubeCarrier installation and prevent the same `Region` being registered in multiple KubeCarrier `Peers`.

`RegionClaim` objects are Namespaced and belong to Providers.

`ServiceCluster` objects are stuck in a `Pending` state until their `RegionClaim` could is bound and thus acknowledged by the central `KubeCarrier` installation.

```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: RegionClaim
metadata:
  name: eu-west-1
  namespace: loodse
spec: {}
status:
  phase: Bound
  observedGeneration: 0
---
apiVersion: kubecarrier.io/v1alpha1
kind: RegionClaim
metadata:
  name: us-east-1
  namespace: loodse
spec: {}
status:
  phase: Conflict
  observedGeneration: 0
```

### Peers

Peers are KubeCarrier installations that are deployed in another region. Multiple Peers can be controlled from a single "Master" and the KubeCarrier API will provide access to all Peers.

Each `Peer` is responsible for a set of Regions the existence of multiple Peers is hidden from Tenants. Providers can spread management of their Services across multiple `Peers` for resiliency and because of network considerations (Latency, Firewalls, etc) e.g. We don't want to manage Services in Japan from a KubeCarrier installation in Germany.

`Peer` objects configures access to other `KubeCarrier` installations for multi region deployments.
A default `local` Peer will be always present with a `KubeCarrier` installation thats set to `master=true`.

`Peer` objects are Cluster-Scoped.

```yaml
apiVersion: master.kubecarrier.io/v1alpha1
kind: Peer
metadata:
  name: local
spec: {}
status:
  phase: Ready
  conditions: []
  observedGeneration: 0
  regions: []
  kubeCarrierVersion:
    version: v0.2.0
    #...
---
apiVersion: master.kubecarrier.io/v1alpha1
kind: Peer
metadata:
  name: eu-west
spec:
  kubeconfigSecret:
    name: eu-west-kubeconfig
status:
  phase: Ready
  conditions: []
  observedGeneration: 0
  regions:
  - name: eu-west-1
    provider: loodse
  - name: eu-west-2
    provider: loodse
  - name: eu-west-1
    provider: not-loodse
---
apiVersion: master.kubecarrier.io/v1alpha1
kind: Peer
metadata:
  name: us-east
spec:
  kubeconfigSecret:
    name: us-east-kubeconfig
status:
  phase: Ready
  conditions: []
  observedGeneration: 0
  regions:
  - name: us-east-1
    provider: loodse
  - name: us-east-1
    provider: not-loodse
```

The `KubeCarrier` object will be extended with a new `master` flag (default False) that will trigger the deployment of the KubeCarrier API server and Multi-Region controllers.

```yaml
apiVersion: operator.kubecarrier.io/v1alpha1
kind: KubeCarrier
metadata:
  name: kubecarrier
spec:
  master: true
```

KubeCarrier installations with the `master` flag set to true, cannot be registered as a `Peer`.

### Account Management

`Accounts` should me managed in the central Master KubeCarrier installation and are reconciled into all `Peers`. `Accounts` in the master cluster will report additional conditions about the state of this reconciliation into multiple `Peers`. `Accounts` in the master cluster are considered `Unready` when the Account is not present in all `Peers`.

## Components

### Master Controller Manager

The Master Controller Manager checks registered `Peer` objects, coordinates `RegionClaims` and `Accounts` across them.

### API

The KubeCarrier API will be the main integration point for external tooling to interact with KubeCarrier.
This contains the CLI, Client-Side Operators, Service Catalog Adapters and UIs.

As the central interface to KubeCarrier, it should hide or abstract complexity as much as possible, making it easy to create integrations on top of it.

Features:
- routes/aggregates requests to/from `Peers`
- aggregates the same type across multiple `Regions`/`ServiceClusters`
  - e.g. exposes `CouchDB.eu-west-1.loodse` and `CouchDB.eu-east-1.loodse` as `CouchDB.loodse`
  - common list and watch interface across types
- RBAC should be offloaded to Kubernetes (if extra RBAC is needed)
- Must retain user-context via e.g. impersonation
- aggregates catalog information from all `Peers`

## Tasks

- Rename `ServiceClusterReference` to `Region`
- TBD
