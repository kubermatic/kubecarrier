---
title: Architecture
weight: 20
date: 2020-04-24T09:00:00+02:00
---

KubeCarrier consists of multiple components that are installed on a single Kubernetes Cluster, referred to as *Management Cluster*.

All components take the form of Kubernetes Controllers working with `CustomResourceDefinitions` and are build using the [kubebuilder project](https://github.com/kubernetes-sigs/kubebuilder).

## Components

### KubeCarrier CLI

The KubeCarrier CLI is a `kubectl` plugin that simplifies the management of your KubeCarrier installation, by providing helpers to validate the environment, trigger the KubeCarrier installation and work with KubeCarrier's APIs.

### KubeCarrier Operator

The KubeCarrier Operator is managing the core KubeCarrier installation and its dynamic components. It runs as a Kubernetes controller and continuously reconciles the KubeCarrier installation to ensure its operation.

### KubeCarrier Manager

The KubeCarrier Manager is the central component of KubeCarrier, that contains all core control loops.

### Ferry

KubeCarrier's `Ferry` component is responsible for managing the connection to a service cluster, which includes health checking, reporting the Kubernetes version and automated setting up of Namespaces in the connected cluster.

### Catapult

A `Catapult` instance is automatically created when a `CustomResourceDiscovery` instance was able to discover a CustomResource from a service cluster and the CRD was successfully established within the management cluster's api machinery.

Each `Catapult` instance is responsible for reconciling one `CustomResourceDefinition` type from the management cluster to a service cluster.

### Elevator

An `Elevator` instance is automatically created when a `DerivedCustomResource` instance established a derived `CustomResourceDefinition`.

Each `Elevator` instance is reconciling one type of `CustomResourceDefinition` to its base.
