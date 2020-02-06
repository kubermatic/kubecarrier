# Dynamic Webhook for arbitrary CRD support

## Introduction

In KubeCarrier, the main feature is to provide services by working with arbitrary CRDs. `Provider` can config
the fields that can be exposed to the `Tenant` by creating the `DerivedCustomResourceDefinition` object. Then the
Tenant can create objects based on it.

In some cases, not all the mandatory fields are exposed to the Tenant within the `DerivedCustomResourceDefinition`.
in this scenario, it is very important to have mutating/validating webhooks to perform necessary defaulting and validating.

## Mutating/Validating Webhooks
Mutating/Validating webhooks are supported by performing a dry-run operation (create/update) against the kube-apiserver, specifically, webhooks will be deployed by catapult and elevator components:
- elevator: performs dry-run operation to against master cluster kube-apiserver (webhooks of Provider-wise CRD (InternalCRD))
- catapult: performs dry-run operation to against service cluster kube-apiserver (webhooks of the baseCRD in the CustomResourceDefinitionDiscovery)

### Request Flow
Tenant creates `couchdbexternals.eu-west-1.example.cloud`
-> master cluster kube-apiserver webhooks
-> dry-run against Internal CRD `CouchdbInternal` webhooks
-> service cluster kube-apiserver webhooks
-> dry-run against Internal CRD `Couchdb` webhooks
