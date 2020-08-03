# [KEP-001] Derived CR support removal 

## Background

Currently, the workflow for propagating custrom resource from the service cluster to the managment cluster's tenant, and back works like this:

* custom resource is 2-way-propagated from service cluster to the management cluster via the ferry. Its group is translated to the managment clusters internal group. (i.e. `internal.svc-1.provider-1`)
* Such exposed custom resource passes through derivation step, one this proposal aims removing. The derivation step serves two purposes:
    * filtering which custom resource fields are exposed
    * exposing it via offering to the tenants
    
The original aim for derivation is lightweight service offering templating and field filtering by the service provider to its consumers, that is kubecarrier tenants. 

For example a database offering could original be like this on the service cluster:

```yaml
apiGroup: db.io
kind: Redis
metadata:
    name: example
    namespace: tenant-1sfsd1
spec:
    instanceSize: large
status:
    url: "redis://example.db.io"
    phase: "Ready"
    conditions: [...]
```

Would be transported via the ferry to the following internal representation:

```yaml
apiGroup: internal.svc-1.provider
kind: Redis
metadata:
    name: example
    namespace: tenant-1sfsd1
spec:
    instanceSize: large
status:
    url: "redis://example.db.io"
    phase: "Ready"
    conditions: [...]
```

and finally via derivation & filtering we get the final resource exposed to the tenant:

```yaml
apiGroup: internal.svc-1.provider
kind: Redis
metadata:
    name: example
    namespace: tenant-1sfsd1
spec: {}
status:
    url: "redis://example.db.io"
    phase: "Ready"
```

Right now no defaulting in the derived custrom resource is supported, only fields filtering.

## Issue

* This process is mandatory, not optional. Thus for exposed services we have to have 2 step dance, including the additional latencies introduced in every layer
* (small) additional maintenance burden for the DerivedCustromResource
* Open issue regarding what the service clusters should do to expose their services:
    * In first scenario they already have CRDs, and their controler which are multi-cluster friendly (i.e. for http services it automatically configures ingress, and exposed said ingress url in the CRD status). This is unlikely scenario.
    * They have to write small, minimal, glue code operator for wrapping existing operators into multi-cluster friendly CRDs. The biggest blocker for generic solution is services & network heterogeneity (e.g. what the service does? How is it exposed? TCP/HTTPS/UDP, how are the secrets/credentials back propagated, etc.)
  
## Suggested changes

* Removing `derivedCustomResource` from the kubecarrier until further use cases are proven useful. 
* In conjunction with potential customers, professional services, and our own technical stewardship evaluate this feature usefulness versus the maintenance cost (i.e. all code must die sometime, or live as zombie forever)