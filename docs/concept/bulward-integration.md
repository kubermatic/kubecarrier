# KubeCarrier -> Bulward Integration

## Introduction
We want to offload permission handling from KubeCarrier to Bulward.

## KubeCarrier

In `KubeCarrier` `Account` represents `Provider` or `Tenant` or both. `Account`s is cluster-scoped objects.
Under the hood `Account` implemented as `Namespace` which determines the scope of visibility for `Tenant` or `Provider`. `Provider`s can discover `Tenant`s and add `CatalogEntries` for them.

## Bulward

In `Bulward` we have only `Organization` and `Project`, which also represented as `Namespace`, but `Organization` is a cluster-scoped object and `Project` namespace-scoped because `Project`s nested in `Organization`.

## Bulward + KubeCarrier


Initial `KubeCarrier` schema:

![kubecarrier](/docs/concept/kubecarrier.svg "KubeCarrier")

Because we don't want to expose `Organization` structure but `Provider` need to have a list of `Tenant`s to share with, Bulward `Organization` can have `Tenant Definition` in it. In that way, `Provider`s will be sharing services with `Organization` and organization owner can configure which `Project` has access to which services.

![tenant](/docs/concept/integration_tenant.svg "Tenant")

`Organization` can have several `Project`s which provides different service, that's mean each `Provider` `Project` should be separate `KubeCarrier` `Account`.

![provider](/docs/concept/integration_provider.svg "Provider")


When `Organization` want's to provide and consume services, `Organization` will be `KubeCarrier` `Account` with `Tenant Definition` and each `Project` that wants to provide service,  will be `KubeCarrier` `Account` with `Provider Definition`

![tenant+provider](/docs/concept/integration_tenant_provider.svg "Tenant + Provider")
