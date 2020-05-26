---
title: KubeCarrier
date: 2020-04-24T09:00:00+02:00
---

![KubeCarrier logo][logo]


# What is KubeCarrier?

KubeCarrier is a open source project for managing services across multiple Kubernetes Clusters running on multiple clouds across multiple regions and provides facilities to provide these services to external users in a self service catalog.

## Why?

**Story time!**

Since the introduction of `CustomResourceDefinitions`, Kubernetes can be easily extended. Allowing you to create extensions of Kubernetes that leverage the same architecture and backend machinery that drive the success of the rest of Kubernetes.

This mechanism is used in the creation of **Operators**: application specific, higher-level extensions of Kubernetes that enable the Kubernetes-native management of more advanced workload. A whole range of Kubernetes Operators can be found on the [operatorhub.io](https://operatorhub.io).

Installing such Kubernetes Operators allows you and your team to focus on your own business and automate the infrastructure management, e.g. Databases on Kubernetes.

All you have to do is to create a Database object and the Operator will do all the heavy lifting for you!

```yaml
apiVersion: dboperator.example.io/v1
kind: Database
metadata:
  name: my-db
spec: {} # DB configuration
```

**Ok... but where does KubeCarrier fit into the picture?**

Glad you asked!

**> Automate the full life cycle**

Now that operators are available and we can automate the whole life cycle of services and applications on Kubernetes, why do we still have to provision, reconfigure and teardown these services via emails and tickets?

**KubeCarrier Catalog** is built to automate the rest of an application life cycle, by giving you as a provider the tools to hand control over to your end users.

So they can create and use your services when they want and you can lean back, drink your coffee and take care of new features of your platform.

**> Keep your overview, across multiple clouds**

As your platform grows, you will also have to manage these services across multiple kubernetes clusters, running on multiple cloud providers, across multiple regions.

Thats where the **KubeCarrier Cross Cluster Manager** is getting into the picture. It helps you to locate, inspect and manage services across your whole platform.
Independent of *Cloud*, *Datacenter* and *Region*.

## How?

KubeCarrier is just yet another Kubernetes Operator, using `CustomResourceDefinitions` and the Kubernetes Controller pattern to do its magic.

Checkout our *Getting Started* docs to see how easy it is to setup and play around with KubeCarrier.

[logo]: ./img/KubeCarrier.png
