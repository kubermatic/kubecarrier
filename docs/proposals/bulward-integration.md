# Bulward Integration
## Introduction
This proposal is about how we integrate [Bulward - Multi Projects/Users Management](https://github.com/kubermatic/bulward)
project to KubeCarrier. This is because bulward has more advanced RBAC management features, and we can offload permission
handling completely from KubeCarrier.
## Objects Reorganization
### KubeCarrier Account
In KubeCarrier, the only object for user representation is `Account`. Here is an example of what `Account` object looks like:
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
  - Tenant
  subjects:
  - kind: User
    name: hannibal
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: team-a-member
    apiGroup: rbac.authorization.k8s.io
```
`roles` in the `Account` indecates if the users in `Account` are service providers or service consumers (tenants). In
KubeCarrier ServiceHub, `Account` can be both `Provider` and `Tenant` at the same time, which means user can consume services
via KubeCarrier ServiceHub, at the same time, user can also provide the services to other users.
`Account` is a **cluster-scoped** object, and each `Account` has a dedicated namespace.
### Bulward Organization & Project
In bulward, there are two objects for user representation.
One is `Organization`, which is cluster-scoped:
```yaml
apiVersion: apiserver.bulward.io/v1alpha1
kind: Organization
metadata:
  name: organization-a
spec:
  metadata:
    displayName: The A Organization
    description: I am the first orgainzation of Bulward.
  owners:
  - kind: User
    name: owner
    apiGroup: rbac.authorization.k8s.io
```
And each `Organization` has a namespace, Organization owner or the user who has permission to create `Project` objects can
create `Project` object in the `Organization` namespace:
```yaml
apiVersion: apiserver.bulward.io/v1alpha1
kind: Project
metadata:
  name: project-example
  namespace: organization-a
spec:
  owners:
  - kind: User
    name: owner
    apiGroup: rbac.authorization.k8s.io
```
`Project` is namespace-scoped, and similar to `Organization`, every `Project` will have a namespace.

### Account <-> Organization or Project?
There are few reasons that we don't replace the `Account` with `Project`:
- KubeCarrier cannot take fully advantage of Organization features in Bulward. For big enterprises, `Organzation` concept
is needed for them to manage permissions of different team, different user has different permission for different project.
- `Account` is cluster-scoped object, while `Project` is namespace-scoped object, integration will introduce lot of changes.

So this proposal proposes the following changes:
We use `Organziation` to represent the `Account` in KubeCarrier. For better explanation, here I will provide an example of
`company-a` wants to consume `couchdb` services via KubeCarrier, and wants to offer `redisdb` to other users.
So we have an `Account` for `company-a`
```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Account
metadata:
  name: company-a
spec:
  metadata:
    displayName: The A Company
    shortDescription: In 1972, a crack commando unit was sent to prison by a military court...
  roles:
  - Provider
  - Tenant
  subjects:
  - kind: User
    name: couchdb-consumer
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: redisdb-operator
    apiGroup: rbac.authorization.k8s.io
```
This can work, but not work nicely, because in KubeCarrier, we cannot config permissions for different user in the same `Account`.
For example, both `couchdb-consumer` and `redisdb-operator` will all have access to both `couchdb` and `redisdb` resources.
A possible workaround would be using two `Account`s for `couchdb-consumer` and `redisdb-operator`, but that just misses the point
to have the option to make an `Account` be Provider and Tenant at the same time.
to solve this issue, we can use `Organization` to represent `Account`:
```yaml
apiVersion: apiserver.bulward.io/v1alpha1
kind: Organization
metadata:
  name: company-a
spec:
  metadata:
    displayName: The A Company
    description: In 1972, a crack commando unit was sent to prison by a military court...
  owners:
  - kind: User
    name: company-owner
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: couchdb-consumer
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: redisdb-operator
    apiGroup: rbac.authorization.k8s.io
```
And then we can have dedicated `Project` for both `couchdb` and `redisdb`:
```yaml
apiVersion: apiserver.bulward.io/v1alpha1
kind: Project
metadata:
  name: couchdb
  namespace: organization-a
spec:
  owners:
  - kind: User
    name: owner
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: couchdb-consumer
    apiGroup: rbac.authorization.k8s.io
---
apiVersion: apiserver.bulward.io/v1alpha1
kind: Project
metadata:
  name: redisdb
  namespace: organization-a
spec:
  owners:
  - kind: User
    name: owner
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: redisdb-operator
    apiGroup: rbac.authorization.k8s.io
```
In this way, we can easily grant different permissions for different team, and different projects. For example, in bulward,
we can use `ProjectRoleTemplate` and `ProjectRole` to set up permission for a (group of) `Project`.

Overall:
- `Account` should be mapped to `Organzation` in a way, not `Project`.
- `Organization` can be `Provider` and `Tenant` at the same time, but `Project` under
`Organization` should be defined as Provider project or Tenant project.

## Implementation
### Deployment
Every bulward release can contain a manifest yaml file, just like cert-manager, we can install it in the similar way:
```bash
kubectl apply -f https://github.com/kubermatic/bulward/releases/download/${BULWARD_VERSION}/bulward.yaml
kubectl wait --for=condition=available deployment/bulward-controller-manager -n bulward-system --timeout=120s
kubectl wait --for=condition=available deployment/bulward-apiserver -n bulward-system --timeout=120s
```
### API Server
#### Organization Endpoint
We have an endpoint for `Account`, we can reuse it for `Organization`, also, in Bulward, only `Organization` object is visible
for organization member (this is implemented in Bulward api extension server), we can also drop the filtering implementation in currently account endpoint.
In addition, we need to inject user impersionation client for this endpoint.
#### Project Endpoint
Project Endpoint (Get/List/Watch/Update/Delete) should be implemented, all permission filtering are offloaded to Bulward api
extension server, so this will also be a thin layer.
In addition, we need to inject user impersionation client for this endpoint.
For filtering Provider/Tenant projects, extra labels can be added to `Project`, so project-admin can have a view of provider projects
and tenant projects in the Organization.

Also extra endpoints may be needed, for example, `OrganizationRole`, `ProjectRole`, etc.

### Controller
#### Max Permission
Since KubeCarrier works with dynamic type of resources, and in bulward, max permission of `Organization` is configured
by `OrganizationRoleTemplate`, KubeCarrier should have permission to update/create `OrganizationRoleTemplate`. All the
permission degradation will be handled by Bulward. Also, organization owner or rbac-admin can set different perssmions
for different project by using `ProjectRole`.
#### Account Replacement
Let's go back to `company-a` example above, `couchdb` and `redisdb` are two projects under the same
`Organization`, and `couchdb` is project to consume couchdb services from other KubeCarrier user, and `redisdb` is a project
to provide `redisdb` services from `company-a` to other users.
In our controllers, we replace the namespace of `Account` with the namespace of `Project`, which means:
For `couchdb`, is a tenant project, so member of this project should be able to work with:
- offerings
- providers
- regions
- coudhdbs resources

For `redisdb`
- catalogs
- catalogentries
- catalogentrysets
- derivedcustomresources,
- redisdb internal resources

For differentiating the Project is a Provider one or Tenant one, we can add labels to `Project`:
```yaml
"kubecarrier.io/project-role": "provider" or "tenant"
```
