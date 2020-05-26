---
title: External And Internal CustomResourceDefinitions
menuTitle: CRDs
weight: 60
pre: <b>5. </b>
slug: crds
date: 2020-04-24T09:00:00+02:00
---

In 4. Catalogs, we created two CRDs. A public one, that users can interact with and an internal one.
This split allows the Provider to override user properties or hide settings and status information from their users.

Now we will create a `CouchDB` instance and see how we work with those objects in KubeCarrier:

**CouchDB definition**

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
