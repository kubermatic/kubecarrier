---
title: API Reference
weight: 50
date: 2020-04-24T09:00:00+02:00
---

The KubeCarrier API is implemented as a extension of the [Kubernetes API](https://kubernetes.io/docs/reference/using-api/api-overview/) as `CustomResourceDefinitions`.
All available objects and their usage are described below.

The API consists of multiple API groups:
* [kubecarrier.io](#core) - Core
* [catalog.kubecarrier.io](#catalog) - Catalog
* [operator.kubecarrier.io](#operator) - Operator

## Core

The core `kubecarrier.io` API group contains the basic buildings blocks of KubeCarrier and objects to setup cross-cluster management of resources.

* [CustomResourceDiscovery.kubecarrier.io/v1alpha1](#customresourcediscoverykubecarrieriov1alpha1)
* [CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1](#customresourcediscoveryconditionkubecarrieriov1alpha1)
* [CustomResourceDiscoveryList.kubecarrier.io/v1alpha1](#customresourcediscoverylistkubecarrieriov1alpha1)
* [CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1](#customresourcediscoveryspeckubecarrieriov1alpha1)
* [CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1](#customresourcediscoverystatuskubecarrieriov1alpha1)
* [CustomResourceDiscoverySet.kubecarrier.io/v1alpha1](#customresourcediscoverysetkubecarrieriov1alpha1)
* [CustomResourceDiscoverySetCRDReference.kubecarrier.io/v1alpha1](#customresourcediscoverysetcrdreferencekubecarrieriov1alpha1)
* [CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1](#customresourcediscoverysetconditionkubecarrieriov1alpha1)
* [CustomResourceDiscoverySetList.kubecarrier.io/v1alpha1](#customresourcediscoverysetlistkubecarrieriov1alpha1)
* [CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1](#customresourcediscoverysetspeckubecarrieriov1alpha1)
* [CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1](#customresourcediscoverysetstatuskubecarrieriov1alpha1)
* [ServiceCluster.kubecarrier.io/v1alpha1](#serviceclusterkubecarrieriov1alpha1)
* [ServiceClusterCondition.kubecarrier.io/v1alpha1](#serviceclusterconditionkubecarrieriov1alpha1)
* [ServiceClusterList.kubecarrier.io/v1alpha1](#serviceclusterlistkubecarrieriov1alpha1)
* [ServiceClusterMetadata.kubecarrier.io/v1alpha1](#serviceclustermetadatakubecarrieriov1alpha1)
* [ServiceClusterSpec.kubecarrier.io/v1alpha1](#serviceclusterspeckubecarrieriov1alpha1)
* [ServiceClusterStatus.kubecarrier.io/v1alpha1](#serviceclusterstatuskubecarrieriov1alpha1)
* [ServiceClusterAssignment.kubecarrier.io/v1alpha1](#serviceclusterassignmentkubecarrieriov1alpha1)
* [ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1](#serviceclusterassignmentconditionkubecarrieriov1alpha1)
* [ServiceClusterAssignmentList.kubecarrier.io/v1alpha1](#serviceclusterassignmentlistkubecarrieriov1alpha1)
* [ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1](#serviceclusterassignmentspeckubecarrieriov1alpha1)
* [ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1](#serviceclusterassignmentstatuskubecarrieriov1alpha1)
* [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1)

### CustomResourceDiscovery.kubecarrier.io/v1alpha1

CustomResourceDiscovery tells KubeCarrier to discover a CustomResource from a ServiceCluster, register it in the Management Cluster and start a new process to reconcile instances of this CRD.

New instances of the CRD will be reconciled by creating a matching instance in the ServiceCluster. Each Namespace in the Managment Cluster needs a ServiceClusterAssignment object, mapping it to a Namespace in the ServiceCluster.

A CustomResourceDiscovery instance will be ready, if the CustomResource was found in the ServiceCluster and a clone of it is established in the Management Cluster. Deleting the instance will also remove the CRD and all instances of it.

**Example**
```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: CustomResourceDiscovery
metadata:
  name: couchdb.eu-west-1
spec:
  crd:
    name: couchdbs.couchdb.io
  serviceCluster:
    name: eu-west-1
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1](#customresourcediscoveryspeckubecarrieriov1alpha1) | false |
| status |  | [CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1](#customresourcediscoverystatuskubecarrieriov1alpha1) | false |

[Back to Group](#core)

### CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoveryCondition contains details of the current state of this CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.kubecarrier.io/v1alpha1 | true |
| type | Type of the condition, currently ('Ready'). | CustomResourceDiscoveryConditionType.kubecarrier.io/v1alpha1 | true |

[Back to Group](#core)

### CustomResourceDiscoveryList.kubecarrier.io/v1alpha1

CustomResourceDiscoveryList is a list of CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][CustomResourceDiscovery.kubecarrier.io/v1alpha1](#customresourcediscoverykubecarrieriov1alpha1) | true |

[Back to Group](#core)

### CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1

CustomResourceDiscoverySpec describes the desired state of a CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |
| serviceCluster | ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |
| kindOverride | KindOverride overrides the kind of the discovered CRD. | string | false |
| webhookStrategy | WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by this CustomResourceDiscovery. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | WebhookStrategyType.kubecarrier.io/v1alpha1 | false |

[Back to Group](#core)

### CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1

CustomResourceDiscoveryStatus represents the observed state of a CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD defines the original CustomResourceDefinition specification from the service cluster. | *apiextensionsv1.CustomResourceDefinition | false |
| managementClusterCRD | ManagementClusterCRD references the CustomResourceDefinition that is created by a CustomResourceDiscovery. | *[ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | CustomResourceDiscoveryPhaseType.kubecarrier.io/v1alpha1 | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | [][CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1](#customresourcediscoveryconditionkubecarrieriov1alpha1) | false |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |

[Back to Group](#core)

### CustomResourceDiscoverySet.kubecarrier.io/v1alpha1

CustomResourceDiscoverySet manages multiple CustomResourceDiscovery objects for a set of ServiceClusters.

**Example**
```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: CustomResourceDiscoverySet
metadata:
  name: couchdb
spec:
  crd:
    name: couchdbs.couchdb.io
  serviceClusterSelector: {}
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1](#customresourcediscoverysetspeckubecarrieriov1alpha1) | false |
| status |  | [CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1](#customresourcediscoverysetstatuskubecarrieriov1alpha1) | false |

[Back to Group](#core)

### CustomResourceDiscoverySetCRDReference.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetCRDReference references a discovered CustomResourceDefinition.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd |  | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |
| serviceCluster |  | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |

[Back to Group](#core)

### CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetCondition contains details for the current condition of this CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.kubecarrier.io/v1alpha1 | true |
| type | Type of the condition, currently ('Ready'). | CustomResourceDiscoverySetConditionType.kubecarrier.io/v1alpha1 | true |

[Back to Group](#core)

### CustomResourceDiscoverySetList.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetList is a list of CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][CustomResourceDiscoverySet.kubecarrier.io/v1alpha1](#customresourcediscoverysetkubecarrieriov1alpha1) | true |

[Back to Group](#core)

### CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetSpec describes the desired state of a CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |
| serviceClusterSelector | ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | true |
| kindOverride | KindOverride overrides the kind of the discovered CRD. | string | false |
| webhookStrategy | WebhookStrategy configs the webhooks of the CRDs which are registered in the management cluster by this CustomResourceDiscoverySet. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | WebhookStrategyType.kubecarrier.io/v1alpha1 | false |

[Back to Group](#core)

### CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetStatus represents the observed state of a CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| managementClusterCRDs | ManagementClusterCRDs contains the CRDs information that created by the CustomResourceDiscovery objects of this CustomResourceDiscoverySet. | [][CustomResourceDiscoverySetCRDReference.kubecarrier.io/v1alpha1](#customresourcediscoverysetcrdreferencekubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | CustomResourceDiscoverySetPhaseType.kubecarrier.io/v1alpha1 | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | [][CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1](#customresourcediscoverysetconditionkubecarrieriov1alpha1) | false |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |

[Back to Group](#core)

### ServiceCluster.kubecarrier.io/v1alpha1

ServiceCluster represents a Kubernets Cluster registered into KubeCarrier.

**Example**
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

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [ServiceClusterSpec.kubecarrier.io/v1alpha1](#serviceclusterspeckubecarrieriov1alpha1) | false |
| status |  | [ServiceClusterStatus.kubecarrier.io/v1alpha1](#serviceclusterstatuskubecarrieriov1alpha1) | false |

[Back to Group](#core)

### ServiceClusterCondition.kubecarrier.io/v1alpha1

ServiceClusterCondition contains details for the current condition of this ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastHeartbeatTime | LastHeartbeatTime is the timestamp corresponding to the last update of this condition. | metav1.Time | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.kubecarrier.io/v1alpha1 | true |
| type | Type of the condition, currently ('Ready'). | ServiceClusterConditionType.kubecarrier.io/v1alpha1 | true |

[Back to Group](#core)

### ServiceClusterList.kubecarrier.io/v1alpha1

ServiceClusterList contains a list of ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][ServiceCluster.kubecarrier.io/v1alpha1](#serviceclusterkubecarrieriov1alpha1) | true |

[Back to Group](#core)

### ServiceClusterMetadata.kubecarrier.io/v1alpha1

ServiceClusterMetadata describes metadata of the ServiceCluster for the Service Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName is the human-readable name of this ServiceCluster. | string | false |
| description | Description is the human-readable description of this ServiceCluster. | string | false |

[Back to Group](#core)

### ServiceClusterSpec.kubecarrier.io/v1alpha1

ServiceClusterSpec describes the desired state of a ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata for display in the Service Catalog. | [ServiceClusterMetadata.kubecarrier.io/v1alpha1](#serviceclustermetadatakubecarrieriov1alpha1) | false |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |

[Back to Group](#core)

### ServiceClusterStatus.kubecarrier.io/v1alpha1

ServiceClusterStatus represents the observed state of a ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | ServiceClusterPhaseType.kubecarrier.io/v1alpha1 | false |
| conditions | Conditions is a list of all conditions this ServiceCluster is in. | [][ServiceClusterCondition.kubecarrier.io/v1alpha1](#serviceclusterconditionkubecarrieriov1alpha1) | false |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |
| kubernetesVersion | KubernetesVersion of the service cluster API Server | *version.Info | false |

[Back to Group](#core)

### ServiceClusterAssignment.kubecarrier.io/v1alpha1

ServiceClusterAssignment is assigning a Namespace in the Management cluster with a Namespace on the ServiceCluster.

The Namespace in the ServiceCluster will be created automatically and is reported in the instance status.

**Example**
```yaml
apiVersion: kubecarrier.io/v1alpha1
kind: ServiceClusterAssignment
metadata:
  name: example1.eu-west-1
spec:
  serviceCluster:
    name: eu-west-1
  managementNamespace:
    name: example1
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1](#serviceclusterassignmentspeckubecarrieriov1alpha1) | false |
| status |  | [ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1](#serviceclusterassignmentstatuskubecarrieriov1alpha1) | false |

[Back to Group](#core)

### ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1

ServiceClusterAssignmentCondition contains details for the current condition of this ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.kubecarrier.io/v1alpha1 | true |
| type | Type of the condition, currently ('Ready'). | ServiceClusterAssignmentConditionType.kubecarrier.io/v1alpha1 | true |

[Back to Group](#core)

### ServiceClusterAssignmentList.kubecarrier.io/v1alpha1

ServiceClusterAssignmentList contains a list of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][ServiceClusterAssignment.kubecarrier.io/v1alpha1](#serviceclusterassignmentkubecarrieriov1alpha1) | true |

[Back to Group](#core)

### ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1

ServiceClusterAssignmentSpec describes the desired state of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| serviceCluster | References the ServiceCluster. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |
| managementNamespace | References the source namespace in the management cluster. | [ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | true |

[Back to Group](#core)

### ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1

ServiceClusterAssignmentStatus represents the observed state of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | ServiceClusterAssignmentPhaseType.kubecarrier.io/v1alpha1 | false |
| conditions | Conditions is a list of all conditions this ServiceClusterAssignment is in. | [][ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1](#serviceclusterassignmentconditionkubecarrieriov1alpha1) | false |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |
| serviceClusterNamespace | ServiceClusterNamespace references the Namespace on the ServiceCluster that was assigned. | *[ObjectReference.kubecarrier.io/v1alpha1](#objectreferencekubecarrieriov1alpha1) | false |

[Back to Group](#core)

### ObjectReference.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same Namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to Group](#core)
## Catalog

The `catalog.kubecarrier.io` API group contains all objects that are used to setup service catalogs in KubeCarrier.

* [Account.catalog.kubecarrier.io/v1alpha1](#accountcatalogkubecarrieriov1alpha1)
* [AccountCondition.catalog.kubecarrier.io/v1alpha1](#accountconditioncatalogkubecarrieriov1alpha1)
* [AccountList.catalog.kubecarrier.io/v1alpha1](#accountlistcatalogkubecarrieriov1alpha1)
* [AccountMetadata.catalog.kubecarrier.io/v1alpha1](#accountmetadatacatalogkubecarrieriov1alpha1)
* [AccountSpec.catalog.kubecarrier.io/v1alpha1](#accountspeccatalogkubecarrieriov1alpha1)
* [AccountStatus.catalog.kubecarrier.io/v1alpha1](#accountstatuscatalogkubecarrieriov1alpha1)
* [Catalog.catalog.kubecarrier.io/v1alpha1](#catalogcatalogkubecarrieriov1alpha1)
* [CatalogCondition.catalog.kubecarrier.io/v1alpha1](#catalogconditioncatalogkubecarrieriov1alpha1)
* [CatalogList.catalog.kubecarrier.io/v1alpha1](#cataloglistcatalogkubecarrieriov1alpha1)
* [CatalogSpec.catalog.kubecarrier.io/v1alpha1](#catalogspeccatalogkubecarrieriov1alpha1)
* [CatalogStatus.catalog.kubecarrier.io/v1alpha1](#catalogstatuscatalogkubecarrieriov1alpha1)
* [CatalogEntry.catalog.kubecarrier.io/v1alpha1](#catalogentrycatalogkubecarrieriov1alpha1)
* [CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1](#catalogentryconditioncatalogkubecarrieriov1alpha1)
* [CatalogEntryList.catalog.kubecarrier.io/v1alpha1](#catalogentrylistcatalogkubecarrieriov1alpha1)
* [CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrymetadatacatalogkubecarrieriov1alpha1)
* [CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1](#catalogentryspeccatalogkubecarrieriov1alpha1)
* [CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrystatuscatalogkubecarrieriov1alpha1)
* [DerivedConfig.catalog.kubecarrier.io/v1alpha1](#derivedconfigcatalogkubecarrieriov1alpha1)
* [CatalogEntrySet.catalog.kubecarrier.io/v1alpha1](#catalogentrysetcatalogkubecarrieriov1alpha1)
* [CatalogEntrySetCondition.catalog.kubecarrier.io/v1alpha1](#catalogentrysetconditioncatalogkubecarrieriov1alpha1)
* [CatalogEntrySetList.catalog.kubecarrier.io/v1alpha1](#catalogentrysetlistcatalogkubecarrieriov1alpha1)
* [CatalogEntrySetMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrysetmetadatacatalogkubecarrieriov1alpha1)
* [CatalogEntrySetSpec.catalog.kubecarrier.io/v1alpha1](#catalogentrysetspeccatalogkubecarrieriov1alpha1)
* [CatalogEntrySetStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrysetstatuscatalogkubecarrieriov1alpha1)
* [CustomResourceDiscoverySetConfig.catalog.kubecarrier.io/v1alpha1](#customresourcediscoverysetconfigcatalogkubecarrieriov1alpha1)
* [CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformationcatalogkubecarrieriov1alpha1)
* [CRDVersion.catalog.kubecarrier.io/v1alpha1](#crdversioncatalogkubecarrieriov1alpha1)
* [DerivedCustomResource.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcecatalogkubecarrieriov1alpha1)
* [DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourceconditioncatalogkubecarrieriov1alpha1)
* [DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcelistcatalogkubecarrieriov1alpha1)
* [DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcespeccatalogkubecarrieriov1alpha1)
* [DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcestatuscatalogkubecarrieriov1alpha1)
* [FieldPath.catalog.kubecarrier.io/v1alpha1](#fieldpathcatalogkubecarrieriov1alpha1)
* [VersionExposeConfig.catalog.kubecarrier.io/v1alpha1](#versionexposeconfigcatalogkubecarrieriov1alpha1)
* [Offering.catalog.kubecarrier.io/v1alpha1](#offeringcatalogkubecarrieriov1alpha1)
* [OfferingList.catalog.kubecarrier.io/v1alpha1](#offeringlistcatalogkubecarrieriov1alpha1)
* [OfferingMetadata.catalog.kubecarrier.io/v1alpha1](#offeringmetadatacatalogkubecarrieriov1alpha1)
* [OfferingSpec.catalog.kubecarrier.io/v1alpha1](#offeringspeccatalogkubecarrieriov1alpha1)
* [Provider.catalog.kubecarrier.io/v1alpha1](#providercatalogkubecarrieriov1alpha1)
* [ProviderList.catalog.kubecarrier.io/v1alpha1](#providerlistcatalogkubecarrieriov1alpha1)
* [ProviderSpec.catalog.kubecarrier.io/v1alpha1](#providerspeccatalogkubecarrieriov1alpha1)
* [Region.catalog.kubecarrier.io/v1alpha1](#regioncatalogkubecarrieriov1alpha1)
* [RegionList.catalog.kubecarrier.io/v1alpha1](#regionlistcatalogkubecarrieriov1alpha1)
* [RegionSpec.catalog.kubecarrier.io/v1alpha1](#regionspeccatalogkubecarrieriov1alpha1)
* [Tenant.catalog.kubecarrier.io/v1alpha1](#tenantcatalogkubecarrieriov1alpha1)
* [TenantList.catalog.kubecarrier.io/v1alpha1](#tenantlistcatalogkubecarrieriov1alpha1)
* [TenantSpec.catalog.kubecarrier.io/v1alpha1](#tenantspeccatalogkubecarrieriov1alpha1)
* [CommonMetadata.catalog.kubecarrier.io/v1alpha1](#commonmetadatacatalogkubecarrieriov1alpha1)
* [Image.catalog.kubecarrier.io/v1alpha1](#imagecatalogkubecarrieriov1alpha1)
* [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1)

### Account.catalog.kubecarrier.io/v1alpha1

Account represents an actor in KubeCarrier. Depending on it's roles, it can provide services, consume offered services or both.

KubeCarrier creates a new Namespace for each Account. The Account Metadata is exposed to users that are offered services from this Account.

**Example**
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
  - Tenant
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [AccountSpec.catalog.kubecarrier.io/v1alpha1](#accountspeccatalogkubecarrieriov1alpha1) | false |
| status |  | [AccountStatus.catalog.kubecarrier.io/v1alpha1](#accountstatuscatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### AccountCondition.catalog.kubecarrier.io/v1alpha1

AccountCondition contains details for the current condition of this Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Account condition, currently ('Ready'). | AccountConditionType.catalog.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.catalog.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### AccountList.catalog.kubecarrier.io/v1alpha1

AccountList contains a list of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Account.catalog.kubecarrier.io/v1alpha1](#accountcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### AccountMetadata.catalog.kubecarrier.io/v1alpha1

AccountMetadata contains the metadata of the Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### AccountSpec.catalog.kubecarrier.io/v1alpha1

AccountSpec describes the desired state of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata\tcontains additional human readable account details. | [AccountMetadata.catalog.kubecarrier.io/v1alpha1](#accountmetadatacatalogkubecarrieriov1alpha1) | false |
| roles | Roles this account uses. | []AccountRole.catalog.kubecarrier.io/v1alpha1 | true |
| subjects | Subjects holds references to the objects that manged RBAC roles should apply to. | []rbacv1.Subject | true |

[Back to Group](#catalog)

### AccountStatus.catalog.kubecarrier.io/v1alpha1

AccountStatus represents the observed state of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | NamespaceName is the name of the Namespace that the Account manages. | *[ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Account by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a Account's current state. | [][AccountCondition.catalog.kubecarrier.io/v1alpha1](#accountconditioncatalogkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | AccountPhaseType.catalog.kubecarrier.io/v1alpha1 | false |

[Back to Group](#catalog)

### Catalog.catalog.kubecarrier.io/v1alpha1

Catalog publishes a selection of CatalogEntries to a selection of Tenants.

KubeCarrier will automatically create ServiceClusterAssignment objects for each Tenant selected by the Catalog.

**Example**
```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: Catalog
metadata:
  name: default
spec:
  tenantSelector: {}
  catalogEntrySelector: {}
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CatalogSpec.catalog.kubecarrier.io/v1alpha1](#catalogspeccatalogkubecarrieriov1alpha1) | false |
| status |  | [CatalogStatus.catalog.kubecarrier.io/v1alpha1](#catalogstatuscatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### CatalogCondition.catalog.kubecarrier.io/v1alpha1

CatalogCondition contains details for the current condition of this Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catalog condition, currently ('Ready'). | CatalogConditionType.catalog.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.catalog.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogList.catalog.kubecarrier.io/v1alpha1

CatalogList contains a list of Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Catalog.catalog.kubecarrier.io/v1alpha1](#catalogcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CatalogSpec.catalog.kubecarrier.io/v1alpha1

CatalogSpec describes the desired contents of a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| catalogEntrySelector | CatalogEntrySelector selects CatalogEntry objects that should be part of this catalog. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |
| tenantSelector | TenantSelector selects Tenant objects that the catalog should be published to. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |

[Back to Group](#catalog)

### CatalogStatus.catalog.kubecarrier.io/v1alpha1

CatalogStatus represents the observed state of Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tenants | Tenants is the list of the Tenants(Tenant) that selected by this Catalog. | [][ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | false |
| entries | Entries is the list of the CatalogEntries that selected by this Catalog. | [][ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catalog by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a Catalog's current state. | [][CatalogCondition.catalog.kubecarrier.io/v1alpha1](#catalogconditioncatalogkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | CatalogPhaseType.catalog.kubecarrier.io/v1alpha1 | false |

[Back to Group](#catalog)

### CatalogEntry.catalog.kubecarrier.io/v1alpha1

CatalogEntry controls how to offer a CRD to other Tenants.

A CatalogEntry references a single CRD, adds metadata to it and allows to limit field access for Tenants.

**Simple Example**
```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: CatalogEntry
metadata:
  name: couchdbs.eu-west-1
spec:
  metadata:
    displayName: CouchDB
    description: The compfy database
  baseCRD:
    name: couchdbs.eu-west-1.loodse
```

**Example with limited fields**
```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: CatalogEntry
metadata:
  name: couchdbs.eu-west-1
spec:
  metadata:
    displayName: CouchDB
    description: The compfy database
  baseCRD:
    name: couchdbs.eu-west-1.loodse
  derive:
    kindOverride: CouchDBPublic
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

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1](#catalogentryspeccatalogkubecarrieriov1alpha1) | false |
| status |  | [CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrystatuscatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1

CatalogEntryCondition contains details for the current condition of this CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntry condition, currently ('Ready'). | CatalogEntryConditionType.catalog.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.catalog.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogEntryList.catalog.kubecarrier.io/v1alpha1

CatalogEntryList contains a list of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][CatalogEntry.catalog.kubecarrier.io/v1alpha1](#catalogentrycatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1

CatalogEntryMetadata contains metadata of the CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1

CatalogEntrySpec describes the desired state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata of the CatalogEntry for the Service Catalog. | [CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrymetadatacatalogkubecarrieriov1alpha1) | true |
| baseCRD | BaseCRD is the underlying BaseCRD objects that this CatalogEntry refers to. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |
| derive | Derive contains the configuration to generate DerivedCustomResource from the BaseCRD of this CatalogEntry. | *[DerivedConfig.catalog.kubecarrier.io/v1alpha1](#derivedconfigcatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1

CatalogEntryStatus represents the observed state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tenantCRD | TenantCRD holds the information about the Tenant facing CRD that is offered by this CatalogEntry. | *[CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformationcatalogkubecarrieriov1alpha1) | false |
| providerCRD | ProviderCRD holds the information about the Provider facing CRD that is offered by this CatalogEntry. | *[CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformationcatalogkubecarrieriov1alpha1) | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntry by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntry's current state. | [][CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1](#catalogentryconditioncatalogkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | CatalogEntryPhaseType.catalog.kubecarrier.io/v1alpha1 | false |

[Back to Group](#catalog)

### DerivedConfig.catalog.kubecarrier.io/v1alpha1

DerivedConfig can be used to limit fields that should be exposed to a Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | [][VersionExposeConfig.catalog.kubecarrier.io/v1alpha1](#versionexposeconfigcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CatalogEntrySet.catalog.kubecarrier.io/v1alpha1

CatalogEntrySet manages a CustomResourceDiscoverySet and creates CatalogEntries for each CRD discovered from the selected ServiceClusters.

**Example**
See CatalogEntry documentation for more configuration details.
```yaml
apiVersion: catalog.kubecarrier.io/v1alpha1
kind: CatalogEntrySet
metadata:
  name: couchdbs
spec:
  metadata:
    displayName: CouchDB
    description: The compfy database
  discoverySet:
    crd:
      name: couchdbs.couchdb.io
    serviceClusterSelector: {}
```

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CatalogEntrySetSpec.catalog.kubecarrier.io/v1alpha1](#catalogentrysetspeccatalogkubecarrieriov1alpha1) | false |
| status |  | [CatalogEntrySetStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrysetstatuscatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### CatalogEntrySetCondition.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetCondition contains details for the current condition of this CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntrySet condition, currently ('Ready'). | CatalogEntrySetConditionType.catalog.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.catalog.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogEntrySetList.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetList contains a list of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][CatalogEntrySet.catalog.kubecarrier.io/v1alpha1](#catalogentrysetcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CatalogEntrySetMetadata.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetMetadata contains the metadata (display name, description, etc) of the CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### CatalogEntrySetSpec.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetSpec defines the desired state of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata of each CatalogEntry for the Service Catalog. | [CatalogEntrySetMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrysetmetadatacatalogkubecarrieriov1alpha1) | true |
| derive | Derive contains the configuration to generate DerivedCustomResources from the BaseCRDs that are selected by this CatalogEntrySet. | *[DerivedConfig.catalog.kubecarrier.io/v1alpha1](#derivedconfigcatalogkubecarrieriov1alpha1) | false |
| discover | Discover contains the configuration to create a CustomResourceDiscoverySet. | [CustomResourceDiscoverySetConfig.catalog.kubecarrier.io/v1alpha1](#customresourcediscoverysetconfigcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CatalogEntrySetStatus.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetStatus defines the observed state of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntrySet by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntrySet's current state. | [][CatalogEntrySetCondition.catalog.kubecarrier.io/v1alpha1](#catalogentrysetconditioncatalogkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | CatalogEntrySetPhaseType.catalog.kubecarrier.io/v1alpha1 | false |

[Back to Group](#catalog)

### CustomResourceDiscoverySetConfig.catalog.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |
| serviceClusterSelector | ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | true |
| kindOverride | KindOverride overrides resulting internal CRDs kind | string | false |
| webhookStrategy | WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by CustomResourceDiscovery object. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | corev1alpha1.WebhookStrategyType | false |

[Back to Group](#catalog)

### CRDInformation.catalog.kubecarrier.io/v1alpha1

CRDInformation contains type information about the CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| apiGroup |  | string | true |
| kind |  | string | true |
| plural |  | string | true |
| versions |  | [][CRDVersion.catalog.kubecarrier.io/v1alpha1](#crdversioncatalogkubecarrieriov1alpha1) | true |
| region | Region references a Region of this CRD. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### CRDVersion.catalog.kubecarrier.io/v1alpha1

CRDVersion holds CRD version specific details.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of this version, for example: v1, v1alpha1, v1beta1 | string | true |
| schema | Schema of this CRD version. | *apiextensionsv1.CustomResourceValidation | false |

[Back to Group](#catalog)

### DerivedCustomResource.catalog.kubecarrier.io/v1alpha1

DerivedCustomResource derives a new CRD from a existing one.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcespeccatalogkubecarrieriov1alpha1) | false |
| status |  | [DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcestatuscatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceCondition contains details for the current condition of this DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the DerivedCustomResource condition, currently ('Ready'). | DerivedCustomResourceConditionType.catalog.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.catalog.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceList contains a list of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][DerivedCustomResource.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcecatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceSpec defines the desired state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| baseCRD | CRD that should be used as a base to derive a new CRD from. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | [][VersionExposeConfig.catalog.kubecarrier.io/v1alpha1](#versionexposeconfigcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceStatus defines the observed state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this DerivedCustomResource by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a DerivedCustomResource's current state. | [][DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourceconditioncatalogkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | DerivedCustomResourcePhaseType.catalog.kubecarrier.io/v1alpha1 | false |
| derivedCR | DerivedCR holds information about the derived CRD. | *[ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### FieldPath.catalog.kubecarrier.io/v1alpha1

FieldPath is specifying how to address a certain field.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jsonPath | JSONPath e.g. .spec.somefield.somesubfield | string | true |

[Back to Group](#catalog)

### VersionExposeConfig.catalog.kubecarrier.io/v1alpha1

VersionExposeConfig specifies which fields to expose in the derived CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| versions | specifies the versions of the referenced CRD, that this expose config applies to. The same version may not be specified in multiple VersionExposeConfigs. | []string | true |
| fields | specifies the fields that should be present in the derived CRD. | [][FieldPath.catalog.kubecarrier.io/v1alpha1](#fieldpathcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### Offering.catalog.kubecarrier.io/v1alpha1

Offering is used for Tenants to discover services that have been made available to them.

Offering objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [OfferingSpec.catalog.kubecarrier.io/v1alpha1](#offeringspeccatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### OfferingList.catalog.kubecarrier.io/v1alpha1

OfferingList contains a list of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Offering.catalog.kubecarrier.io/v1alpha1](#offeringcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### OfferingMetadata.catalog.kubecarrier.io/v1alpha1

OfferingMetadata contains the metadata (display name, description, etc) of the Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### OfferingSpec.catalog.kubecarrier.io/v1alpha1

OfferingSpec defines the data (metadata, provider, crds, etc.) of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [OfferingMetadata.catalog.kubecarrier.io/v1alpha1](#offeringmetadatacatalogkubecarrieriov1alpha1) | true |
| provider | Provider references the Provider managing this Offering. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |
| crd | CRD holds the information about the underlying CRD that is offered by this offering. | [CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformationcatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### Provider.catalog.kubecarrier.io/v1alpha1

Provider exposes information of an Account with the Provider role.

Provider objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [ProviderSpec.catalog.kubecarrier.io/v1alpha1](#providerspeccatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### ProviderList.catalog.kubecarrier.io/v1alpha1

ProviderList contains a list of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Provider.catalog.kubecarrier.io/v1alpha1](#providercatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### ProviderSpec.catalog.kubecarrier.io/v1alpha1

ProviderSpec defines the desired state of Provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the Provider. | [AccountMetadata.catalog.kubecarrier.io/v1alpha1](#accountmetadatacatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### Region.catalog.kubecarrier.io/v1alpha1

Region exposes information about a Providers Cluster.

Region objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [RegionSpec.catalog.kubecarrier.io/v1alpha1](#regionspeccatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### RegionList.catalog.kubecarrier.io/v1alpha1

RegionList contains a list of Region.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Region.catalog.kubecarrier.io/v1alpha1](#regioncatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### RegionSpec.catalog.kubecarrier.io/v1alpha1

RegionSpec defines the desired state of Region

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the ServiceCluster. | corev1alpha1.ServiceClusterMetadata | false |
| provider | Provider references the Provider that this ServiceCluster belongs to. | [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreferencecatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### Tenant.catalog.kubecarrier.io/v1alpha1

Tenant exposes information about available Tenants on the platform and allows a Provider to set custom labels on them.

Tenant objects will be created for all Accounts with the role \"Tenant\" in all Account Namespaces with the role \"Provider\".

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [TenantSpec.catalog.kubecarrier.io/v1alpha1](#tenantspeccatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### TenantList.catalog.kubecarrier.io/v1alpha1

TenantList contains a list of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Tenant.catalog.kubecarrier.io/v1alpha1](#tenantcatalogkubecarrieriov1alpha1) | true |

[Back to Group](#catalog)

### TenantSpec.catalog.kubecarrier.io/v1alpha1

TenantSpec defines the desired state of Tenant

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### CommonMetadata.catalog.kubecarrier.io/v1alpha1

CommonMetadata contains human-readable information shared for all catalog related objects.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName is the human-readable name of this Service. | string | true |
| description | Description is the long and detailed description of the Service. | string | false |
| shortDescription | ShortDescription is a single line short description of the Service. | string | true |
| logo | Logo is the full sized logo of the service. | *[Image.catalog.kubecarrier.io/v1alpha1](#imagecatalogkubecarrieriov1alpha1) | false |
| icon | Icon is a small squared logo of the service. | *[Image.catalog.kubecarrier.io/v1alpha1](#imagecatalogkubecarrieriov1alpha1) | false |

[Back to Group](#catalog)

### Image.catalog.kubecarrier.io/v1alpha1

Image describes an inlined image.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| mediaType | MediaType of the included image in data. e.g. image/png, image/jpeg, image/svg | string | true |
| data | Data is the image data. | []byte.catalog.kubecarrier.io/v1alpha1 | true |

[Back to Group](#catalog)

### ObjectReference.catalog.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to Group](#catalog)
## Operator

The `operator.kubecarrier.io` API group contains objects to interact with the KubeCarrier installation.

* [APIServer.operator.kubecarrier.io/v1alpha1](#apiserveroperatorkubecarrieriov1alpha1)
* [APIServerCondition.operator.kubecarrier.io/v1alpha1](#apiserverconditionoperatorkubecarrieriov1alpha1)
* [APIServerList.operator.kubecarrier.io/v1alpha1](#apiserverlistoperatorkubecarrieriov1alpha1)
* [APIServerOIDCConfig.operator.kubecarrier.io/v1alpha1](#apiserveroidcconfigoperatorkubecarrieriov1alpha1)
* [APIServerSpec.operator.kubecarrier.io/v1alpha1](#apiserverspecoperatorkubecarrieriov1alpha1)
* [APIServerStatus.operator.kubecarrier.io/v1alpha1](#apiserverstatusoperatorkubecarrieriov1alpha1)
* [StaticUsers.operator.kubecarrier.io/v1alpha1](#staticusersoperatorkubecarrieriov1alpha1)
* [Catapult.operator.kubecarrier.io/v1alpha1](#catapultoperatorkubecarrieriov1alpha1)
* [CatapultCondition.operator.kubecarrier.io/v1alpha1](#catapultconditionoperatorkubecarrieriov1alpha1)
* [CatapultList.operator.kubecarrier.io/v1alpha1](#catapultlistoperatorkubecarrieriov1alpha1)
* [CatapultSpec.operator.kubecarrier.io/v1alpha1](#catapultspecoperatorkubecarrieriov1alpha1)
* [CatapultStatus.operator.kubecarrier.io/v1alpha1](#catapultstatusoperatorkubecarrieriov1alpha1)
* [Elevator.operator.kubecarrier.io/v1alpha1](#elevatoroperatorkubecarrieriov1alpha1)
* [ElevatorCondition.operator.kubecarrier.io/v1alpha1](#elevatorconditionoperatorkubecarrieriov1alpha1)
* [ElevatorList.operator.kubecarrier.io/v1alpha1](#elevatorlistoperatorkubecarrieriov1alpha1)
* [ElevatorSpec.operator.kubecarrier.io/v1alpha1](#elevatorspecoperatorkubecarrieriov1alpha1)
* [ElevatorStatus.operator.kubecarrier.io/v1alpha1](#elevatorstatusoperatorkubecarrieriov1alpha1)
* [Ferry.operator.kubecarrier.io/v1alpha1](#ferryoperatorkubecarrieriov1alpha1)
* [FerryCondition.operator.kubecarrier.io/v1alpha1](#ferryconditionoperatorkubecarrieriov1alpha1)
* [FerryList.operator.kubecarrier.io/v1alpha1](#ferrylistoperatorkubecarrieriov1alpha1)
* [FerrySpec.operator.kubecarrier.io/v1alpha1](#ferryspecoperatorkubecarrieriov1alpha1)
* [FerryStatus.operator.kubecarrier.io/v1alpha1](#ferrystatusoperatorkubecarrieriov1alpha1)
* [KubeCarrier.operator.kubecarrier.io/v1alpha1](#kubecarrieroperatorkubecarrieriov1alpha1)
* [KubeCarrierCondition.operator.kubecarrier.io/v1alpha1](#kubecarrierconditionoperatorkubecarrieriov1alpha1)
* [KubeCarrierList.operator.kubecarrier.io/v1alpha1](#kubecarrierlistoperatorkubecarrieriov1alpha1)
* [KubeCarrierSpec.operator.kubecarrier.io/v1alpha1](#kubecarrierspecoperatorkubecarrieriov1alpha1)
* [KubeCarrierStatus.operator.kubecarrier.io/v1alpha1](#kubecarrierstatusoperatorkubecarrieriov1alpha1)
* [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreferenceoperatorkubecarrieriov1alpha1)
* [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1)

### APIServer.operator.kubecarrier.io/v1alpha1

APIServer manages the deployment of the KubeCarrier central API server.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [APIServerSpec.operator.kubecarrier.io/v1alpha1](#apiserverspecoperatorkubecarrieriov1alpha1) | false |
| status |  | [APIServerStatus.operator.kubecarrier.io/v1alpha1](#apiserverstatusoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### APIServerCondition.operator.kubecarrier.io/v1alpha1

APIServerCondition contains details for the current condition of this APIServer.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the APIServer condition, currently ('Ready'). | APIServerConditionType.operator.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.operator.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### APIServerList.operator.kubecarrier.io/v1alpha1

APIServerList contains a list of APIServer

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][APIServer.operator.kubecarrier.io/v1alpha1](#apiserveroperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### APIServerOIDCConfig.operator.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| issuerURL | IssuerURL is the URL the provider signs ID Tokens as. This will be the \"iss\" field of all tokens produced by the provider and is used for configuration discovery.\n\nThe URL is usually the provider's URL without a path, for example \"https://accounts.google.com\" or \"https://login.salesforce.com\".\n\nThe provider must implement configuration discovery. See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfig | string | true |
| clientID | ClientID the JWT must be issued for, the \"sub\" field. This plugin only trusts a single client to ensure the plugin can be used with public providers.\n\nThe plugin supports the \"authorized party\" OpenID Connect claim, which allows specialized providers to issue tokens to a client for a different client. See: https://openid.net/specs/openid-connect-core-1_0.html#IDToken | string | true |
| apiAudiences | APIAudiences are the audiences that the API server identitifes as. The (API audiences unioned with the ClientIDs) should have a non-empty intersection with the request's target audience. This preserves the behavior of the OIDC authenticator pre-introduction of API audiences. | authenticator.Audiences | false |
| certificateAuthority | CertificateAuthority references the secret containing issuer's CA in a PEM encoded root certificate of the provider. | [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | true |
| usernameClaim | UsernameClaim is the JWT field to use as the user's username. | string | true |
| usernamePrefix | UsernamePrefix, if specified, causes claims mapping to username to be prefix with the provided value. A value \"oidc:\" would result in usernames like \"oidc:john\". | string | false |
| groupsClaim | GroupsClaim, if specified, causes the OIDCAuthenticator to try to populate the user's groups with an ID Token field. If the GroupsClaim field is present in an ID Token the value must be a string or list of strings. | string | false |
| groupsPrefix | GroupsPrefix, if specified, causes claims mapping to group names to be prefixed with the value. A value \"oidc:\" would result in groups like \"oidc:engineering\" and \"oidc:marketing\". | string | false |
| supportedSigningAlgs | SupportedSigningAlgs sets the accepted set of JOSE signing algorithms that can be used by the provider to sign tokens.\n\nhttps://tools.ietf.org/html/rfc7518#section-3.1\n\nThis value defaults to RS256, the value recommended by the OpenID Connect spec:\n\nhttps://openid.net/specs/openid-connect-core-1_0.html#IDTokenValidation | []string | false |
| requiredClaims | RequiredClaims, if specified, causes the OIDCAuthenticator to verify that all the required claims key value pairs are present in the ID Token. | map[string]string | false |

[Back to Group](#operator)

### APIServerSpec.operator.kubecarrier.io/v1alpha1

APIServerSpec defines the desired state of APIServer

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tlsSecretRef | TLSSecretRef references the TLS certificate and private key for serving the KubeCarrier API. | *[ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | false |
| oidc | OIDC specifies OpenID Connect configuration for API Server authentication | *[APIServerOIDCConfig.operator.kubecarrier.io/v1alpha1](#apiserveroidcconfigoperatorkubecarrieriov1alpha1) | false |
| staticUsers | StaticUsers specifies static users configuration for API Server authentication | *[StaticUsers.operator.kubecarrier.io/v1alpha1](#staticusersoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### APIServerStatus.operator.kubecarrier.io/v1alpha1

APIServerStatus defines the observed state of APIServer

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this APIServer by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a APIServer's current state. | [][APIServerCondition.operator.kubecarrier.io/v1alpha1](#apiserverconditionoperatorkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | APIServerPhaseType.operator.kubecarrier.io/v1alpha1 | false |

[Back to Group](#operator)

### StaticUsers.operator.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| htpasswdSecret | HtpassswdSecret specifies the htpasswd secret to use for static user authentication. | [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### Catapult.operator.kubecarrier.io/v1alpha1

Catapult manages the deployment of the Catapult controller manager.

A Catapult instance is started for each CustomResourceDiscovery instance and responsible for reconciling CRD instances across Kubernetes Clusters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [CatapultSpec.operator.kubecarrier.io/v1alpha1](#catapultspecoperatorkubecarrieriov1alpha1) | false |
| status |  | [CatapultStatus.operator.kubecarrier.io/v1alpha1](#catapultstatusoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### CatapultCondition.operator.kubecarrier.io/v1alpha1

CatapultCondition contains details for the current condition of this Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catapult condition, currently ('Ready'). | CatapultConditionType.operator.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.operator.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### CatapultList.operator.kubecarrier.io/v1alpha1

CatapultList contains a list of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Catapult.operator.kubecarrier.io/v1alpha1](#catapultoperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### CatapultSpec.operator.kubecarrier.io/v1alpha1

CatapultSpec defines the desired state of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| managementClusterCRD | References the CRD in the Management Cluster. | [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreferenceoperatorkubecarrieriov1alpha1) | true |
| serviceClusterCRD | References the CRD in the ServiceCluster. | [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreferenceoperatorkubecarrieriov1alpha1) | true |
| serviceCluster | References the ServiceCluster object that this object belongs to. | [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | true |
| webhookStrategy | WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by this Catapult. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | corev1alpha1.WebhookStrategyType | false |

[Back to Group](#operator)

### CatapultStatus.operator.kubecarrier.io/v1alpha1

CatapultStatus defines the observed state of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catapult by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a Catapult's current state. | [][CatapultCondition.operator.kubecarrier.io/v1alpha1](#catapultconditionoperatorkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | CatapultPhaseType.operator.kubecarrier.io/v1alpha1 | false |

[Back to Group](#operator)

### Elevator.operator.kubecarrier.io/v1alpha1

Elevator manages the deployment of the Elevator controller manager.

For each `DerivedCustomResource` a Elevator instance is launched to propagate the derived CRD instance into the Namespace of it's provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [ElevatorSpec.operator.kubecarrier.io/v1alpha1](#elevatorspecoperatorkubecarrieriov1alpha1) | false |
| status |  | [ElevatorStatus.operator.kubecarrier.io/v1alpha1](#elevatorstatusoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### ElevatorCondition.operator.kubecarrier.io/v1alpha1

ElevatorCondition contains details for the current condition of this Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Elevator condition, currently ('Ready'). | ElevatorConditionType.operator.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.operator.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### ElevatorList.operator.kubecarrier.io/v1alpha1

ElevatorList contains a list of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Elevator.operator.kubecarrier.io/v1alpha1](#elevatoroperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### ElevatorSpec.operator.kubecarrier.io/v1alpha1

ElevatorSpec defines the desired state of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| providerCRD | References the provider or internal CRD, that should be created in the provider namespace. | [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreferenceoperatorkubecarrieriov1alpha1) | true |
| tenantCRD | References the public CRD that will be synced into the provider namespace. | [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreferenceoperatorkubecarrieriov1alpha1) | true |
| derivedCR | References the DerivedCustomResource controlling the Tenant-side CRD. | [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### ElevatorStatus.operator.kubecarrier.io/v1alpha1

ElevatorStatus defines the observed state of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Elevator by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a Elevator's current state. | [][ElevatorCondition.operator.kubecarrier.io/v1alpha1](#elevatorconditionoperatorkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | ElevatorPhaseType.operator.kubecarrier.io/v1alpha1 | false |

[Back to Group](#operator)

### Ferry.operator.kubecarrier.io/v1alpha1

Ferry manages the deployment of the Ferry controller manager.

Ferry lives in the Provider Namespace. For each ferry the KubeCarrier operator spins up
the ferry controller deployment, necessary roles, service accounts, and role bindings.

The reason for ferry controller deployment are multiples:
* security --> KubeCarrier operator has greater privileges then ferry controller
* resource isolation --> each ferry controller pod operates only on a single service cluster,
		thus resource allocation and monitoring is separate per ferry. This allows finer grade
		resource tuning and monitoring
* flexibility --> If needed different ferries could have different deployments depending on
		their specific need (e.g. KubeCarrier image version for gradual rolling upgrade, different resource allocation, etc),

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [FerrySpec.operator.kubecarrier.io/v1alpha1](#ferryspecoperatorkubecarrieriov1alpha1) | false |
| status |  | [FerryStatus.operator.kubecarrier.io/v1alpha1](#ferrystatusoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### FerryCondition.operator.kubecarrier.io/v1alpha1

FerryCondition contains details for the current condition of this Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.operator.kubecarrier.io/v1alpha1 | true |
| type | Type of the condition, currently ('Ready'). | FerryConditionType.operator.kubecarrier.io/v1alpha1 | true |

[Back to Group](#operator)

### FerryList.operator.kubecarrier.io/v1alpha1

FerryList contains a list of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][Ferry.operator.kubecarrier.io/v1alpha1](#ferryoperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### FerrySpec.operator.kubecarrier.io/v1alpha1

FerrySpec defines the desired state of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreferenceoperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### FerryStatus.operator.kubecarrier.io/v1alpha1

FerryStatus defines the observed state of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | FerryPhaseType.operator.kubecarrier.io/v1alpha1 | false |
| conditions | Conditions is a list of all conditions this Ferry is in. | [][FerryCondition.operator.kubecarrier.io/v1alpha1](#ferryconditionoperatorkubecarrieriov1alpha1) | false |
| observedGeneration | The most recent generation observed by the controller. | int64 | false |

[Back to Group](#operator)

### KubeCarrier.operator.kubecarrier.io/v1alpha1

KubeCarrier manages the deployment of the KubeCarrier controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [KubeCarrierSpec.operator.kubecarrier.io/v1alpha1](#kubecarrierspecoperatorkubecarrieriov1alpha1) | false |
| status |  | [KubeCarrierStatus.operator.kubecarrier.io/v1alpha1](#kubecarrierstatusoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### KubeCarrierCondition.operator.kubecarrier.io/v1alpha1

KubeCarrierCondition contains details for the current condition of this KubeCarrier.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the KubeCarrier condition, currently ('Ready'). | KubeCarrierConditionType.operator.kubecarrier.io/v1alpha1 | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | ConditionStatus.operator.kubecarrier.io/v1alpha1 | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### KubeCarrierList.operator.kubecarrier.io/v1alpha1

KubeCarrierList contains a list of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][KubeCarrier.operator.kubecarrier.io/v1alpha1](#kubecarrieroperatorkubecarrieriov1alpha1) | true |

[Back to Group](#operator)

### KubeCarrierSpec.operator.kubecarrier.io/v1alpha1

KubeCarrierSpec defines the desired state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| api |  | [APIServerSpec.operator.kubecarrier.io/v1alpha1](#apiserverspecoperatorkubecarrieriov1alpha1) | false |

[Back to Group](#operator)

### KubeCarrierStatus.operator.kubecarrier.io/v1alpha1

KubeCarrierStatus defines the observed state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this KubeCarrier by the controller. | int64 | false |
| conditions | Conditions represents the latest available observations of a KubeCarrier's current state. | [][KubeCarrierCondition.operator.kubecarrier.io/v1alpha1](#kubecarrierconditionoperatorkubecarrieriov1alpha1) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | KubeCarrierPhaseType.operator.kubecarrier.io/v1alpha1 | false |

[Back to Group](#operator)

### CRDReference.operator.kubecarrier.io/v1alpha1

CRDReference references a CustomResourceDefitition.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kind |  | string | true |
| version |  | string | true |
| group |  | string | true |
| plural |  | string | true |

[Back to Group](#operator)

### ObjectReference.operator.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to Group](#operator)
