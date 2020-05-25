---
title: Accounts
weight: 30
pre: <b>2. </b>
slug: accounts
date: 2020-04-24T09:00:00+02:00
---

KubeCarrier manages everything in Accounts. Each Account is separated by its own Namespace and subjects within the Account get RBAC Roles setup and assigned, so they can interact with the System.

To startup of KubeCarrier, we will create two Accounts. The first account `team-a`, will provide services, while `team-b` will be able to consume services.

Each `Account` has a list of subjects, similar to `RoleBinding` objects. These subjects will be setup with admin rights for their namespace.

Accounts with the:
- `Provider` role can register `ServiceCluster`, manage `Catalogs` and organize their services.
- `Tenant` role can create services that were made available to them via `Catalogs` from a `Provider`.

Accounts may be a `Provider` and a `Tenant` at the same time.

**Account definitions:**

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

We will look more at the differences between the *Provider* and *Tenant* roles for accounts in *4. Catalogs*.
