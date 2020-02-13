# What is KubeCarrier?

KubeCarrier is an open source system for managing applications and services across multiple Kubernetes Clusters; providing a framework to centralize the management of services and provide these services with external users in a self service catalog.

## Why?

**Story time!**

With the introduction of `ThirdPartyResources`, and later the much more successfully `CustomResourceDefinitions`, Kubernetes is now very easy to extend.

This drives the creation and adoption of **Operators**, application specific, higher-level extensions of Kubernetes that enable the kubernetes-native management of more advanced workload.

A whole range of Kubernetes Operators can be found on the [operatorhub.io](https://operatorhub.io).

Installing such Kubernetes Operators allows you and your team to focus on your own business and automate the infrastructure management, e.g. Databases on Kubernetes.

All you have todo is to create a Database object and the Operator will do all the heavy lifting for you!

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

**KubeCarrier Catalog** is build to automate the rest of an application life cycle, by giving you as a provider the tools to hand control over to your customers, your end users.

So they can create and use your services when they want and you can lean back, drink your coffee and take care of new features of your platform.

**> Keep your overview, across multiple clouds**

As your platform grows, you will also have to manage these services across multiple kubernetes clusters, running on multiple cloud providers, across multiple regions.

Thats where the **KubeCarrier Cross Cluster Manager** is getting into the picture. It helps you to locate, inspect and manage services across your whole platform.
Independent of *Cloud*, *Datacenter* and *Region*.

## How?

KubeCarrier is just yet another Kubernetes Operator, using `CustomResourceDefinitions` and the Kubernetes Controller pattern to do its magic.

Checkout our *Getting Started* docs to see how easy it is to setup and play around with KubeCarrier.
