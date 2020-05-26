---
title: Catalogs
weight: 50
pre: <b>4. </b>
slug: catalogs
date: 2020-04-24T09:00:00+02:00
---

## Catalog Entries

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

**CatalogEntrySet definition**

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

## Catalogs

Now that we have successfully registered a `CustomResourceDefinition` from another cluster, attached metadata to it and created a "public" interface for other people, we can go ahead and actually offer this `CouchDB` object to other users.

The `CatalogEntrySet` we created in previous step is managing `CatalogEntries` for all `ServiceClusters` that match the given `serviceClusterSelector`.

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

**Catalog definition**

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
