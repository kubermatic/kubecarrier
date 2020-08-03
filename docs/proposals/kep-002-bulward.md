# [KEP-002] Bulward integration

## Background

Currently, we have `Account` concept in kubecarrier, which encompasses single user. Due to bigger interop with the Kubermatic ecosystem, as well as further requirements on the user management we built `bulward`. The account is permeated through the kubecarrier project, as it's central user concept. The API server's proto file specify the account, the accounts come into two possible roles (Provider and Tenant), and adequate RBAC permissions

## Migration plan

Current `Account` concept is closely matched to bulward's `Project`. Thus all references to the account through the kubecarrier solution should be renamed to account. In the following code excerpt you can see similarities and differences. 

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Account
metadata:
  name: team-a
spec:
  metadata:
    displayName: The A Team
    shortDescription: In 1972, a crack commando unit was sent to prison by a military court...
  roles:
  - Provider
  subjects:
  - kind: User
    name: hannibal
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: team-a-member
    apiGroup: rbac.authorization.k8s.io
```

```yaml
apiVersion: apiserver.bulward.io/v1alpha1
kind: Project
metadata:
  name: project-example
  namespace: organization-a
spec:
  owners:
  - kind: User
    name: hannibal
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: team-a-member
    apiGroup: rbac.authorization.k8s.io
```

### Similarities

* both have ownership concept 
* both define a namespace where the owners & memeber operation take place

### Differences

* `Account` is namespace scoped, while the `Project` is namespaces scopes, within the `Organization` namespace. `Organization` itself is a cluster scoped object.
* `Project` object currently has no metadata concept, though it should be added in further bulward version


## Implementation details

### Deployment

The kubecarrier operator shall deploy versioned pinned bulward solution as its dependency. It will generate the necessary manifest, installed them on the cluster, and let the kubernetes complete the installation (i.e. start pods, connect to API Extension server, etc.)

The `statik` manifests are prepared in the bulward repo, where the operator imports them as any other library. It's versioned pinned in the `go.{mod,sum}` file.

### accounts role

Since the account comes in two role flavours, the provider and the tenant, a concept missing from the bulward, we define two labels:

* `"kubecarrier.io/provider": "true"`
* `"kubecarrier.io/tenant": "true"`

with the same semantics as the accounts `.spec.roles` field. For complete implementation the additional `Project` webhooks needs to be installed, verifying at least one role has been successfully set.

### API Server

Currently, API server performs all operations with account's information. Since there's a difference between cluster scoped `Account` and namespace scope `Project` there are two ways how can we migrate the API server:

* in the object meta proto rename account to the `projectNamespace`, as how was it used internally. 
```protobuf
message ObjectMeta {
  string projectNamespace = 3;
}
```

* have project & organization separated
```protobuf
message ObjectMeta {
  string project = 3;
  string organization = 15;  
}
```

Additional API endpoints need to be added for Project listing (within an organization, and optionally for all organization the user have access to). The bulward API extension server shall handle the permission filtering, with the API server impersonating the user. Previously we haven't impersonated the user due to lacking support within the `controller-runtime` client. This is an open investigation topic.

### Controllers

Most controllers operate on Account by fetching its namespace. That is, things are namespace bound. This is equivalent how the project shall be utilized. The `ProjectRoleTemplate` shall be used instead of creating roles for each user. 

For now all provider projects shall see all tenant project, that is have tenant reference within their namespace. The catalog subsystem shall work similarly as before. 

Controllers shall operate on the `storage.bulward.io` API groups for projects, thus bypassing the view restriction put in place by the bulward extension API server

