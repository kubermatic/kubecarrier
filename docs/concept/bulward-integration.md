# KubeCarrier -> Bulward Integration

## Introduction
We want to offload permission handling from KubeCarrier to to Bulward.

## KubeCarrier

In `KubeCarrier` `Account` represents `Provider` or `Tenant` or both. `Account`s is cluster-scoped objects.
Under the hood `Account` implemented as `Namespace` which determines scope of visiblilty for `Tenant` or `Provider`. `Provider`s can discover `Tenant`s and add `CatalogEntries` for them.

## Bulward

In `Bulward` we have only `Organization` and `Project`, which also represented as `Namespace`, but `Organization` if cluster-scoped object and `Project` namespace-scoped because `Project`s nested in `Organization`.

## Bulward + KubeCarrier

Initial `KubeCarrier` schema:

![kubecarrier](/docs/concept/kubecarrier.svg "KubeCarrier")

Bulward with `Tenant` organization

![tenant](/docs/concept/integration_tenant.svg "Tenant")

Bulward with `Provider` organization

![provider](/docs/concept/integration_provider.svg "Provider")

Bulward with `Provider` + `Provider` organization

![tenant+provider](/docs/concept/integration_tenant_provider.svg "Tenant + Provider")
