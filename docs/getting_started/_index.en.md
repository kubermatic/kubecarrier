---
title: Getting Started
weight: 100
---

## 0. Requirements

To install KubeCarrier you will need a Kubernetes Cluster with the [cert-manager](https://cert-manager.io/docs/) installed.

|Component   |Version       |
|------------|--------------|
|Kubernetes  | v1.16, v1.17 |
|cert-manager| v0.13.0      |

### Getting a Kubernetes Cluster

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

## 1. Install KubeCarrier

KubeCarrier is distributed via a public container registry [quay.io/kubecarrier](https://quay.io/kubecarrier). While the KubeCarrier installation is managed by the KubeCarrier operator, installing and upgrading the operator is done via our kubectl plugin.

This CLI tool will gain more utility functions as the project matures.

### Install the kubectl plugin

To install the kubectl plugin, just visit the KubeCarrier [release page](https://github.com/kubermatic/kubecarrier/releases), download the archive and put the contained `kubecarrier` binary into your `$PATH` as `kubectl-kubecarrier`.

Make sure the binary is executable.

If everything worked, you should now be setup with the `kubecarrier` plugin:
*(Your version should be way newer than this example)*

```bash
$ kubectl kubecarrier version --full
branch: master
buildTime: "2020-02-25T14:03:31Z"
commit: a23bdbe
goVersion: go1.13
platform: linux/amd64
version: master-a23bdbe
```

### Install KubeCarrier

```bash
# make sure you are connected to the cluster,
# that you want to install KubeCarrier on
$ kubectl config current-context
kind-kubecarrier

# install KubeCarrier
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
0.19s ✔  Deploy KubeCarrier Operator
6.29s ✔  Deploy KubeCarrier
```

The `kubectl kubecarrier setup` command is idempotent, so its safe to just re-run it multiple times, if you encounter any error in your setup.

<details>
<summary><b>Debug install issues</b></summary>

KubeCarrier is installed into the `kubecarrier-system` Namespace by default.

If a step in the installation is timing out, you should check the logs of the respective component:

#### Operator
```bash
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
10.09s ✖  Deploy KubeCarrier Operator
Error: deploying kubecarrier operator: timed out waiting for the condition

$ kubectl get po -n kubecarrier-system
NAME                                          READY   STATUS   RESTARTS   AGE
kubecarrier-operator-manager-7d4b8f74-mgbgn   0/1     Error    2          32s

$ kubectl logs -n kubecarrier-system kubecarrier-operator-manager-7d4b8f74-mgbgn
[...]
Error: running manager: no matches for kind "Issuer" in version "cert-manager.io/v1alpha2"
[...]
```

In this case the cert-manager was not installed beforehand.

#### KubeCarrier Control Plane
```bash
$ kubectl kubecarrier setup
0.03s ✔  Create "kubecarrier-system" Namespace
0.19s ✔  Deploy KubeCarrier Operator
60.09s ✖  Deploy KubeCarrier
Error: deploying kubecarrier: timed out waiting for the condition

$ kubectl get po -n kubecarrier-system
NAME                                                      READY   STATUS             RESTARTS   AGE
kubecarrier-manager-controller-manager-56bfd4dcbd-8rg4l   1/1     CrashLoopBackOff   0          11m
kubecarrier-operator-manager-7d4b8f74-vfsxl               1/1     Running            0          11m

$ kubectl logs -n kubecarrier-system kubecarrier-manager-controller-manager-56bfd4dcbd-8rg4l
```

</details>

## 2. Accounts

KubeCarrier manages everything in Accounts. Each Account is separated by its own Namespace and subjects within the Account get RBAC Roles setup and assigned, so they can interact with the System.

To startup of KubeCarrier, we will create two Accounts. The first account `team-a`, will provide services, while `team-b` will be able to consume services.

Each `Account` has a list of subjects, similar to `RoleBinding` objects. These subjects will be setup with admin rights for their namespace.

Accounts with the:
- `Provider` role can register `ServiceCluster`, manage `Catalogs` and organize their services.
- `Tenant` role can create services that were made available to them via `Catalogs` from a `Provider`.

Accounts may be a `Provider` and a `Tenant` at the same time.

<details>
<summary><b>Account definitions</b></summary>

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Account
metadata:
  name: team-a
spec:
  metadata:
    displayName: The A Team
    description: In 1972, a crack commando unit was sent to prison by a military court...
  roles:
  - Provider
  subjects:
  - kind: User
    name: hannibal
    apiGroup: rbac.authorization.k8s.io
  - kind: User
    name: team-a-member
    apiGroup: rbac.authorization.k8s.io
---
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Account
metadata:
  name: team-b
spec:
  roles:
  - Tenant
  subjects:
  - kind: User
    name: team-b-member
    apiGroup: rbac.authorization.k8s.io
```

</details>

To create these objects just run:

```bash
$ kubectl apply \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/accounts.yaml
account.catalog.kubecarrier.io/team-a created
account.catalog.kubecarrier.io/team-b created
```

After creating those accounts, you can check their statuses and namespace:

```bash
$ kubectl get account
NAME     ACCOUNT NAMESPACE   DISPLAY NAME   STATUS   AGE
team-a   team-a              The A Team     Ready    7s
team-b   team-b                             Ready    7s
```

We will look more at the differences between the *Provider* and *Tenant* roles for accounts in [5. Catalogs](#5-catalogs).

## 3. Clusters

Next we want to register Kubernetes Clusters into KubeCarrier.
To begin you need another Kubeconfig.

<details>
<summary><b><i>Need another Cluster?</i></b></summary>
<br>

If you don't have another Kubernetes Cluster, just go back to [0. Requirements](#0-requirements) and create another cluster with Kind.
In this example we will use the name `eu-west-1` for this new cluster.

When you create another cluster with Kind, you have to work with the **internal** Kubeconfig of the cluster, see command below:

`kind get kubeconfig --internal --name eu-west-1 > /tmp/eu-west-1-kubeconfig`

This will replace the default `localhost:xxxx` address with the container's IP address, allowing KubeCarrier to talk with the other kind cluster.

**Attention**
When creating a new cluster with `kind` your active context will be switched to the newly created cluster.
Check `kubectl config current-context` and use `kubectl config use-context` to switch back to the right cluster.

</details>

To begin, we have to upload our Kubeconfig as a `Secret` into our Account Namespace.

```bash
$ kubectl create secret generic eu-west-1-kubeconfig \
  -n team-a \
  --from-file=kubeconfig=/tmp/eu-west-1-kubeconfig
secret/eu-west-1-kubeconfig created
```

Now that we have the credentials and connection information, we can register the Cluster into KubeCarrier.

<details>
<summary><b>ServiceCluster definitions</b></summary>

```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: ServiceCluster
metadata:
  name: eu-west-1
spec:
  metadata:
    displayName: EU West 1
  kubeconfigSecret:
    name: eu-west-1-kubeconfig
```
</details>

Create the object with:

```bash
$ kubectl apply -n team-a \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/servicecluster.yaml
servicecluster.kubecarrier.io/team-a created

$ kubectl get servicecluster -n team-a
NAME        STATUS   DISPLAY NAME   KUBERNETES VERSION   AGE
eu-west-1   Ready    EU West 1      v1.17.0              8s
```

KubeCarrier will connect to the Cluster, do basic health checking and report the Kubernetes Version.

## 4. Catalog Entries

In order to manage Custom Resources from a `ServiceCluster` we have to tell KubeCarrier how to find them and how we want to offer them to our users.

First we need some kind of `CustomResourceDefinition` or Operator installation in our `ServiceCluster`.
To help get you started we have a fictional example CRD that can be used without having to setup an Operator.

Register the CRD in the `ServiceCluster`.

```bash
# make sure you are connected to the ServiceCluster
# thats `eu-west-1` if you followed our earlier guide.
$ kubectl config use-context kind-eu-west-1
Switched to context "kind-eu-west-1".

$ kubectl apply \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/couchdb.crd.yaml
customresourcedefinition.apiextensions.k8s.io/couchdbs.couchdb.io created

$ kubectl get crd
NAME                  CREATED AT
couchdbs.couchdb.io   2020-03-10T10:27:51Z
```

Now we will tell the KubeCarrier installation to work with this CRD.
We can accomplish this, by creating a `CatalogEntrySet`. This object describes which CRD should be fetched from which ServiceCluster, metadata for the Service Catalog and it will limit which fields are available to users.

<details>
<summary><b>CatalogEntrySet definition</b></summary>

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: CatalogEntrySet
metadata:
  name: couchdbs.eu-west-1
spec:
  metadata:
    displayName: CouchDB
    description: The compfy database
  discover:
    crd:
      name: couchdbs.couchdb.io
    serviceClusterSelector: {}
    kindOverride: CouchDBInternal
  derive:
    kindOverride: CouchDB
    expose:
    - versions:
      - v1alpha1
      fields:
      - jsonPath: .spec.username
      - jsonPath: .spec.password
      - jsonPath: .status.phase
      - jsonPath: .status.fauxtonAddress
      - jsonPath: .status.address
      - jsonPath: .status.observedGeneration
```
</details>

```bash
# make sure you are connected to the KubeCarrier Cluster
# thats `kubecarrier` if you followed our earlier guide.
$ kubectl config use-context kind-kubecarrier
Switched to context "kind-kubecarrier".

$ kubectl apply -n team-a \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/catalogentryset.yaml
catalogentryset.catalog.kubecarrier.io/couchdbs created

$ kubectl get catalogentryset -n team-a
NAME       STATUS   CRD                   AGE
couchdbs   Ready    couchdbs.couchdb.io   19s
```

As soon as the `CatalogEntrySet` is ready, you will notice two new `CustomResourceDefinitions` appearing in the Cluster:

```bash
$ kubectl get crd -l kubecarrier.io/origin-namespace=team-a
NAME                                CREATED AT
couchdbinternals.eu-west-1.team-a   2020-03-09T10:28:39Z
couchdbs.eu-west-1.team-a           2020-03-09T10:28:52Z
```

The `couchdbinternals.eu-west-1.team-a` object is just a copy of the CRD present in the `ServiceCluster`, while `couchdbs.eu-west-1.team-a` is a "slimed-down" version, only containing fields specified in the `CatalogEntrySet`. Both CRDs are "namespaced" by their API group.


## 5. Catalogs

Now that we have successfully registered a `CustomResourceDefinition` from another cluster, attached metadata to it and created a "public" interface for other people, we can go ahead and actually offer this `CouchDB` object to other users.

The `CatalogEntrySet` we created in [4. Catalog Entries](#4-catalog-entries) is managing `CatalogEntries` for all `ServiceClusters` that match the given `serviceClusterSelector`.

```bash
$ kubectl get catalogentry -n team-a
NAME                 STATUS   BASE CRD                            TENANT CRD                  AGE
couchdbs.eu-west-1   Ready    couchdbinternals.eu-west-1.team-a   couchdbs.eu-west-1.team-a   26s
```

We can now reference these `CatalogEntries` in a `Catalog` and offer them to `Tenants`.
Every `Account` with the `Tenant` role has a `Tenant` object created in each `Provider` namespace.

```bash
$ kubectl get tenant -n team-a
NAME     AGE
team-b   5m35s
```

These objects allow the `Provider` to organize them by setting labels on them, so they can be selected by a `Catalog`.
This `Catalog` selects all `CatalogEntries` and offers them to all `Tenants`:

<details>
<summary><b>Catalog definition</b></summary>

```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Catalog
metadata:
  name: default
spec:
  # selects all the Tenants
  tenantSelector: {}
  # selects all the CatalogEntries
  catalogEntrySelector: {}
```
</details>

```bash
$ kubectl apply -n team-a \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/catalog.yaml
catalog.catalog.kubecarrier.io/default created

$ kubectl get catalog -n team-a
NAME      STATUS   AGE
default   Ready    5s
```

When the `Catalog` is ready, selected `Tenants` can discover objects available to them and RBAC is setup to users to work with the CRD in their namespace.
Here we also use `kubectl` user impersonation (`--as`), to showcase RBAC:

```bash
# Offering objects contain information about CRDs that are shared to a Tenant.
# They contain all the information to validate and create new instances.
$ kubectl get offering -n team-b --as=team-b-member
NAME                        DISPLAY NAME   PROVIDER   AGE
couchdbs.eu-west-1.team-a   CouchDB        team-a     3m15s

# Region exposes information about the underlying Clusters.
$ kubectl get region -n team-b --as=team-b-member
NAME               PROVIDER   DISPLAY NAME   AGE
eu-west-1.team-a   team-a     EU West 1      5m14s

# Provider exposes information about the Provider of an Offering.
$ kubectl get provider -n team-b --as=team-b-member
NAME     DISPLAY NAME   AGE
team-a   The A Team     6m11s
```

## 6. Internal and External CRD

In [4. Catalog Entries](4-catalog-entries), we created two CRDs. A public one, that users can interact with and an internal one.
This split allows the Provider to override user properties or hide settings and status information from their users.

Now we will create a `CouchDB` instance and see how we work with those objects in KubeCarrier:

<details>
<summary><b>CouchDB definition</b></summary>

```yaml
apiVersion: eu-west-1.team-a/v1alpha1
kind: CouchDB
metadata:
  name: db1
spec:
  username: hans
  password: hans2000
```
</details>

```bash
$ kubectl apply -n team-b --as=team-b-member \
  -f https://raw.githubusercontent.com/kubermatic/kubecarrier/master/docs/manifests/couchdb.eu-west-1.yaml
catalog.catalog.kubecarrier.io/default created

# This is the object we just created as team-b
$ kubectl get couchdb -n team-b --as=team-b-member
NAME   USERNAME   PASSWORD   AGE
db1    hans       hans2000   4s

# There is also an internal representation, that users cannot access:
$ kubectl get couchdbinternal -n team-b --as=team-b-member
Error from server (Forbidden): couchdbinternals.eu-west-1.team-a is forbidden: User "team-b-member" cannot list resource "couchdbinternals" in API group "eu-west-1.team-a" in the namespace "team-b"

# Only members of the Provider team (team-a) can access these objects:
kubectl get couchdbinternal -n team-b --as=team-a-member
NAME   USERNAME   PASSWORD   VERSION   AGE
db1    hans       hans2000             31s
```

**Ok, what is happening here?**

Team A is offering the `CouchDB` service from their Kubernetes cluster `eu-west-1` and Team B created an instance of the `CouchDB` service.

Because Team A decided to hide the `.spec.version` property it's absent from the CRD that tenants of Team A have access to. While the internal CRD retains that field, so the provider can use it to orchestrate their workload.
