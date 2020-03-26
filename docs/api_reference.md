# KubeCarrier - API Reference

The KubeCarrier API is implemented as a extension of the [Kubernetes API](https://kubernetes.io/docs/reference/using-api/api-overview/) as `CustomResourceDefinitions`.
All available objects and their usage are described below.

The API consists of multiple API groups:
* [kubecarrier.io](#core) - Core
* [catalog.kubecarrier.io](#catalog) - Catalog
* [operator.kubecarrier.io](#operator) - Operator

## Core

The core `kubecarrier.io` API group contains the basic buildings blocks of KubeCarrier and objects to setup cross-cluster management of resources.

* [CustomResourceDiscovery.kubecarrier.io/v1alpha1](#customresourcediscovery.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1](#customresourcediscoverycondition.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryList.kubecarrier.io/v1alpha1](#customresourcediscoverylist.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1](#customresourcediscoveryspec.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1](#customresourcediscoverystatus.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySet.kubecarrier.io/v1alpha1](#customresourcediscoveryset.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetCRDReference.kubecarrier.io/v1alpha1](#customresourcediscoverysetcrdreference.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1](#customresourcediscoverysetcondition.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetList.kubecarrier.io/v1alpha1](#customresourcediscoverysetlist.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1](#customresourcediscoverysetspec.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1](#customresourcediscoverysetstatus.kubecarrier.io/v1alpha1)
* [ServiceCluster.kubecarrier.io/v1alpha1](#servicecluster.kubecarrier.io/v1alpha1)
* [ServiceClusterCondition.kubecarrier.io/v1alpha1](#serviceclustercondition.kubecarrier.io/v1alpha1)
* [ServiceClusterList.kubecarrier.io/v1alpha1](#serviceclusterlist.kubecarrier.io/v1alpha1)
* [ServiceClusterMetadata.kubecarrier.io/v1alpha1](#serviceclustermetadata.kubecarrier.io/v1alpha1)
* [ServiceClusterSpec.kubecarrier.io/v1alpha1](#serviceclusterspec.kubecarrier.io/v1alpha1)
* [ServiceClusterStatus.kubecarrier.io/v1alpha1](#serviceclusterstatus.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignment.kubecarrier.io/v1alpha1](#serviceclusterassignment.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1](#serviceclusterassignmentcondition.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentList.kubecarrier.io/v1alpha1](#serviceclusterassignmentlist.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1](#serviceclusterassignmentspec.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1](#serviceclusterassignmentstatus.kubecarrier.io/v1alpha1)
* [ObjectReference.kubecarrier.io/v1alpha1](#objectreference.kubecarrier.io/v1alpha1)

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
| spec |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySpec | false |
| status |  | kubecarrier.io/v1alpha1.CustomResourceDiscoveryStatus | false |

[Back to Group](#core)

### CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoveryCondition contains details of the current state of this CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.CustomResourceDiscoveryConditionType | true |

[Back to Group](#core)

### CustomResourceDiscoveryList.kubecarrier.io/v1alpha1

CustomResourceDiscoveryList is a list of CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.CustomResourceDiscovery | true |

[Back to Group](#core)

### CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1

CustomResourceDiscoverySpec describes the desired state of a CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| serviceCluster | ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on. | kubecarrier.io/v1alpha1.ObjectReference | true |
| kindOverride | KindOverride overrides the kind of the discovered CRD. | string | false |
| webhookStrategy | WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by this CustomResourceDiscovery. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | kubecarrier.io/v1alpha1.WebhookStrategyType | false |

[Back to Group](#core)

### CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1

CustomResourceDiscoveryStatus represents the observed state of a CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD defines the original CustomResourceDefinition specification from the service cluster. | *apiextensionsv1.CustomResourceDefinition | false |
| managementClusterCRD | ManagementClusterCRD references the CustomResourceDefinition that is created by a CustomResourceDiscovery. | *kubecarrier.io/v1alpha1.ObjectReference | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.CustomResourceDiscoveryPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | []kubecarrier.io/v1alpha1.CustomResourceDiscoveryCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |

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
| spec |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetSpec | false |
| status |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetStatus | false |

[Back to Group](#core)

### CustomResourceDiscoverySetCRDReference.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetCRDReference references a discovered CustomResourceDefinition.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd |  | kubecarrier.io/v1alpha1.ObjectReference | true |
| serviceCluster |  | kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#core)

### CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetCondition contains details for the current condition of this CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetConditionType | true |

[Back to Group](#core)

### CustomResourceDiscoverySetList.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetList is a list of CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.CustomResourceDiscoverySet | true |

[Back to Group](#core)

### CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetSpec describes the desired state of a CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| serviceClusterSelector | ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | true |
| kindOverride | KindOverride overrides the kind of the discovered CRD. | string | false |
| webhookStrategy | WebhookStrategy configs the webhooks of the CRDs which are registered in the management cluster by this CustomResourceDiscoverySet. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | kubecarrier.io/v1alpha1.WebhookStrategyType | false |

[Back to Group](#core)

### CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetStatus represents the observed state of a CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| managementClusterCRDs | ManagementClusterCRDs contains the CRDs information that created by the CustomResourceDiscovery objects of this CustomResourceDiscoverySet. | []kubecarrier.io/v1alpha1.CustomResourceDiscoverySetCRDReference | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | []kubecarrier.io/v1alpha1.CustomResourceDiscoverySetCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |

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
| spec |  | kubecarrier.io/v1alpha1.ServiceClusterSpec | false |
| status |  | kubecarrier.io/v1alpha1.ServiceClusterStatus | false |

[Back to Group](#core)

### ServiceClusterCondition.kubecarrier.io/v1alpha1

ServiceClusterCondition contains details for the current condition of this ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastHeartbeatTime | LastHeartbeatTime is the timestamp corresponding to the last update of this condition. | metav1.Time | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.ServiceClusterConditionType | true |

[Back to Group](#core)

### ServiceClusterList.kubecarrier.io/v1alpha1

ServiceClusterList contains a list of ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.ServiceCluster | true |

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
| metadata | Metadata for display in the Service Catalog. | kubecarrier.io/v1alpha1.ServiceClusterMetadata | false |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#core)

### ServiceClusterStatus.kubecarrier.io/v1alpha1

ServiceClusterStatus represents the observed state of a ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.ServiceClusterPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceCluster is in. | []kubecarrier.io/v1alpha1.ServiceClusterCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |
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
| spec |  | kubecarrier.io/v1alpha1.ServiceClusterAssignmentSpec | false |
| status |  | kubecarrier.io/v1alpha1.ServiceClusterAssignmentStatus | false |

[Back to Group](#core)

### ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1

ServiceClusterAssignmentCondition contains details for the current condition of this ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.ServiceClusterAssignmentConditionType | true |

[Back to Group](#core)

### ServiceClusterAssignmentList.kubecarrier.io/v1alpha1

ServiceClusterAssignmentList contains a list of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.ServiceClusterAssignment | true |

[Back to Group](#core)

### ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1

ServiceClusterAssignmentSpec describes the desired state of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| serviceCluster | References the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| managementNamespace | References the source namespace in the management cluster. | kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#core)

### ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1

ServiceClusterAssignmentStatus represents the observed state of ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.ServiceClusterAssignmentPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceClusterAssignment is in. | []kubecarrier.io/v1alpha1.ServiceClusterAssignmentCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |
| serviceClusterNamespace | ServiceClusterNamespace references the Namespace on the ServiceCluster that was assigned. | *kubecarrier.io/v1alpha1.ObjectReference | false |

[Back to Group](#core)

### ObjectReference.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same Namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to Group](#core)
## Catalog

The `catalog.kubecarrier.io` API group contains all objects that are used to setup service catalogs in KubeCarrier.

* [Account.catalog.kubecarrier.io/v1alpha1](#account.catalog.kubecarrier.io/v1alpha1)
* [AccountCondition.catalog.kubecarrier.io/v1alpha1](#accountcondition.catalog.kubecarrier.io/v1alpha1)
* [AccountList.catalog.kubecarrier.io/v1alpha1](#accountlist.catalog.kubecarrier.io/v1alpha1)
* [AccountMetadata.catalog.kubecarrier.io/v1alpha1](#accountmetadata.catalog.kubecarrier.io/v1alpha1)
* [AccountSpec.catalog.kubecarrier.io/v1alpha1](#accountspec.catalog.kubecarrier.io/v1alpha1)
* [AccountStatus.catalog.kubecarrier.io/v1alpha1](#accountstatus.catalog.kubecarrier.io/v1alpha1)
* [Catalog.catalog.kubecarrier.io/v1alpha1](#catalog.catalog.kubecarrier.io/v1alpha1)
* [CatalogCondition.catalog.kubecarrier.io/v1alpha1](#catalogcondition.catalog.kubecarrier.io/v1alpha1)
* [CatalogList.catalog.kubecarrier.io/v1alpha1](#cataloglist.catalog.kubecarrier.io/v1alpha1)
* [CatalogSpec.catalog.kubecarrier.io/v1alpha1](#catalogspec.catalog.kubecarrier.io/v1alpha1)
* [CatalogStatus.catalog.kubecarrier.io/v1alpha1](#catalogstatus.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntry.catalog.kubecarrier.io/v1alpha1](#catalogentry.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1](#catalogentrycondition.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryList.catalog.kubecarrier.io/v1alpha1](#catalogentrylist.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrymetadata.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1](#catalogentryspec.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrystatus.catalog.kubecarrier.io/v1alpha1)
* [DerivedConfig.catalog.kubecarrier.io/v1alpha1](#derivedconfig.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySet.catalog.kubecarrier.io/v1alpha1](#catalogentryset.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySetCondition.catalog.kubecarrier.io/v1alpha1](#catalogentrysetcondition.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySetList.catalog.kubecarrier.io/v1alpha1](#catalogentrysetlist.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySetMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrysetmetadata.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySetSpec.catalog.kubecarrier.io/v1alpha1](#catalogentrysetspec.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySetStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrysetstatus.catalog.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySetConfig.catalog.kubecarrier.io/v1alpha1](#customresourcediscoverysetconfig.catalog.kubecarrier.io/v1alpha1)
* [CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformation.catalog.kubecarrier.io/v1alpha1)
* [CRDVersion.catalog.kubecarrier.io/v1alpha1](#crdversion.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResource.catalog.kubecarrier.io/v1alpha1](#derivedcustomresource.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcecondition.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcelist.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcespec.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcestatus.catalog.kubecarrier.io/v1alpha1)
* [FieldPath.catalog.kubecarrier.io/v1alpha1](#fieldpath.catalog.kubecarrier.io/v1alpha1)
* [VersionExposeConfig.catalog.kubecarrier.io/v1alpha1](#versionexposeconfig.catalog.kubecarrier.io/v1alpha1)
* [Offering.catalog.kubecarrier.io/v1alpha1](#offering.catalog.kubecarrier.io/v1alpha1)
* [OfferingList.catalog.kubecarrier.io/v1alpha1](#offeringlist.catalog.kubecarrier.io/v1alpha1)
* [OfferingMetadata.catalog.kubecarrier.io/v1alpha1](#offeringmetadata.catalog.kubecarrier.io/v1alpha1)
* [OfferingSpec.catalog.kubecarrier.io/v1alpha1](#offeringspec.catalog.kubecarrier.io/v1alpha1)
* [Provider.catalog.kubecarrier.io/v1alpha1](#provider.catalog.kubecarrier.io/v1alpha1)
* [ProviderList.catalog.kubecarrier.io/v1alpha1](#providerlist.catalog.kubecarrier.io/v1alpha1)
* [ProviderSpec.catalog.kubecarrier.io/v1alpha1](#providerspec.catalog.kubecarrier.io/v1alpha1)
* [Region.catalog.kubecarrier.io/v1alpha1](#region.catalog.kubecarrier.io/v1alpha1)
* [RegionList.catalog.kubecarrier.io/v1alpha1](#regionlist.catalog.kubecarrier.io/v1alpha1)
* [RegionSpec.catalog.kubecarrier.io/v1alpha1](#regionspec.catalog.kubecarrier.io/v1alpha1)
* [Tenant.catalog.kubecarrier.io/v1alpha1](#tenant.catalog.kubecarrier.io/v1alpha1)
* [TenantList.catalog.kubecarrier.io/v1alpha1](#tenantlist.catalog.kubecarrier.io/v1alpha1)
* [TenantSpec.catalog.kubecarrier.io/v1alpha1](#tenantspec.catalog.kubecarrier.io/v1alpha1)
* [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreference.catalog.kubecarrier.io/v1alpha1)

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
| spec |  | catalog.kubecarrier.io/v1alpha1.AccountSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.AccountStatus | false |

[Back to Group](#catalog)

### AccountCondition.catalog.kubecarrier.io/v1alpha1

AccountCondition contains details for the current condition of this Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Account condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.AccountConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### AccountList.catalog.kubecarrier.io/v1alpha1

AccountList contains a list of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Account | true |

[Back to Group](#catalog)

### AccountMetadata.catalog.kubecarrier.io/v1alpha1

AccountMetadata contains the metadata of the Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName is the human-readable name of this Account. | string | false |
| description | Description is the human-readable description of this Account. | string | false |

[Back to Group](#catalog)

### AccountSpec.catalog.kubecarrier.io/v1alpha1

AccountSpec describes the desired state of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata\tcontains additional human readable account details. | catalog.kubecarrier.io/v1alpha1.AccountMetadata | false |
| roles | Roles this account uses. | []catalog.kubecarrier.io/v1alpha1.AccountRole | true |
| subjects | Subjects holds references to the objects that manged RBAC roles should apply to. | []rbacv1.Subject | true |

[Back to Group](#catalog)

### AccountStatus.catalog.kubecarrier.io/v1alpha1

AccountStatus represents the observed state of Account.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespace | NamespaceName is the name of the Namespace that the Account manages. | *catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Account by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Account's current state. | []catalog.kubecarrier.io/v1alpha1.AccountCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.AccountPhaseType | false |

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
| spec |  | catalog.kubecarrier.io/v1alpha1.CatalogSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.CatalogStatus | false |

[Back to Group](#catalog)

### CatalogCondition.catalog.kubecarrier.io/v1alpha1

CatalogCondition contains details for the current condition of this Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catalog condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.CatalogConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogList.catalog.kubecarrier.io/v1alpha1

CatalogList contains a list of Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Catalog | true |

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
| tenants | Tenants is the list of the Tenants(Tenant) that selected by this Catalog. | []catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| entries | Entries is the list of the CatalogEntries that selected by this Catalog. | []catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catalog by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catalog's current state. | []catalog.kubecarrier.io/v1alpha1.CatalogCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.CatalogPhaseType | false |

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
| spec |  | catalog.kubecarrier.io/v1alpha1.CatalogEntrySpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.CatalogEntryStatus | false |

[Back to Group](#catalog)

### CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1

CatalogEntryCondition contains details for the current condition of this CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntry condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.CatalogEntryConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogEntryList.catalog.kubecarrier.io/v1alpha1

CatalogEntryList contains a list of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.CatalogEntry | true |

[Back to Group](#catalog)

### CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1

CatalogEntryMetadata contains metadata of the CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this CatalogEntry. | string | true |
| description | Description shows the human-readable description of this CatalogEntry. | string | true |

[Back to Group](#catalog)

### CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1

CatalogEntrySpec describes the desired state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata of the CatalogEntry for the Service Catalog. | catalog.kubecarrier.io/v1alpha1.CatalogEntryMetadata | true |
| baseCRD | BaseCRD is the underlying BaseCRD objects that this CatalogEntry refers to. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
| derive | Derive contains the configuration to generate DerivedCustomResource from the BaseCRD of this CatalogEntry. | *catalog.kubecarrier.io/v1alpha1.DerivedConfig | false |

[Back to Group](#catalog)

### CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1

CatalogEntryStatus represents the observed state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tenantCRD | TenantCRD holds the information about the Tenant facing CRD that is offered by this CatalogEntry. | *catalog.kubecarrier.io/v1alpha1.CRDInformation | false |
| providerCRD | ProviderCRD holds the information about the Provider facing CRD that is offered by this CatalogEntry. | *catalog.kubecarrier.io/v1alpha1.CRDInformation | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntry by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntry's current state. | []catalog.kubecarrier.io/v1alpha1.CatalogEntryCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.CatalogEntryPhaseType | false |

[Back to Group](#catalog)

### DerivedConfig.catalog.kubecarrier.io/v1alpha1

DerivedConfig can be used to limit fields that should be exposed to a Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.VersionExposeConfig | true |

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
| spec |  | catalog.kubecarrier.io/v1alpha1.CatalogEntrySetSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.CatalogEntrySetStatus | false |

[Back to Group](#catalog)

### CatalogEntrySetCondition.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetCondition contains details for the current condition of this CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntrySet condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.CatalogEntrySetConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### CatalogEntrySetList.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetList contains a list of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.CatalogEntrySet | true |

[Back to Group](#catalog)

### CatalogEntrySetMetadata.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetMetadata contains the metadata (display name, description, etc) of the CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this CatalogEntrySet. | string | true |
| description | Description shows the human-readable description of this CatalogEntrySet. | string | true |

[Back to Group](#catalog)

### CatalogEntrySetSpec.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetSpec defines the desired state of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata of each CatalogEntry for the Service Catalog. | catalog.kubecarrier.io/v1alpha1.CatalogEntrySetMetadata | true |
| derive | Derive contains the configuration to generate DerivedCustomResources from the BaseCRDs that are selected by this CatalogEntrySet. | *catalog.kubecarrier.io/v1alpha1.DerivedConfig | false |
| discover | Discover contains the configuration to create a CustomResourceDiscoverySet. | catalog.kubecarrier.io/v1alpha1.CustomResourceDiscoverySetConfig | true |

[Back to Group](#catalog)

### CatalogEntrySetStatus.catalog.kubecarrier.io/v1alpha1

CatalogEntrySetStatus defines the observed state of CatalogEntrySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntrySet by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntrySet's current state. | []catalog.kubecarrier.io/v1alpha1.CatalogEntrySetCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.CatalogEntrySetPhaseType | false |

[Back to Group](#catalog)

### CustomResourceDiscoverySetConfig.catalog.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
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
| versions |  | []catalog.kubecarrier.io/v1alpha1.CRDVersion | true |
| region | Region references a Region of this CRD. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |

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
| spec |  | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceStatus | false |

[Back to Group](#catalog)

### DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceCondition contains details for the current condition of this DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the DerivedCustomResource condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#catalog)

### DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceList contains a list of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.DerivedCustomResource | true |

[Back to Group](#catalog)

### DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceSpec defines the desired state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| baseCRD | CRD that should be used as a base to derive a new CRD from. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.VersionExposeConfig | true |

[Back to Group](#catalog)

### DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceStatus defines the observed state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this DerivedCustomResource by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a DerivedCustomResource's current state. | []catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourcePhaseType | false |
| derivedCR | DerivedCR holds information about the derived CRD. | *catalog.kubecarrier.io/v1alpha1.ObjectReference | false |

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
| fields | specifies the fields that should be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.FieldPath | true |

[Back to Group](#catalog)

### Offering.catalog.kubecarrier.io/v1alpha1

Offering is used for Tenants to discover services that have been made available to them.

Offering objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.OfferingSpec | false |

[Back to Group](#catalog)

### OfferingList.catalog.kubecarrier.io/v1alpha1

OfferingList contains a list of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Offering | true |

[Back to Group](#catalog)

### OfferingMetadata.catalog.kubecarrier.io/v1alpha1

OfferingMetadata contains the metadata (display name, description, etc) of the Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this Offering. | string | true |
| description | Description shows the human-readable description of this Offering. | string | true |

[Back to Group](#catalog)

### OfferingSpec.catalog.kubecarrier.io/v1alpha1

OfferingSpec defines the data (metadata, provider, crds, etc.) of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | catalog.kubecarrier.io/v1alpha1.OfferingMetadata | true |
| provider | Provider references the Provider managing this Offering. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
| crd | CRD holds the information about the underlying CRD that is offered by this offering. | catalog.kubecarrier.io/v1alpha1.CRDInformation | false |

[Back to Group](#catalog)

### Provider.catalog.kubecarrier.io/v1alpha1

Provider exposes information of an Account with the Provider role.

Provider objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.ProviderSpec | false |

[Back to Group](#catalog)

### ProviderList.catalog.kubecarrier.io/v1alpha1

ProviderList contains a list of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Provider | true |

[Back to Group](#catalog)

### ProviderSpec.catalog.kubecarrier.io/v1alpha1

ProviderSpec defines the desired state of Provider

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the Provider. | catalog.kubecarrier.io/v1alpha1.AccountMetadata | true |

[Back to Group](#catalog)

### Region.catalog.kubecarrier.io/v1alpha1

Region exposes information about a Providers Cluster.

Region objects are created automatically by KubeCarrier in Account namespaces, that have a service offered to them via a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.RegionSpec | false |

[Back to Group](#catalog)

### RegionList.catalog.kubecarrier.io/v1alpha1

RegionList contains a list of Region.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Region | true |

[Back to Group](#catalog)

### RegionSpec.catalog.kubecarrier.io/v1alpha1

RegionSpec defines the desired state of Region

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the ServiceCluster. | corev1alpha1.ServiceClusterMetadata | false |
| provider | Provider references the Provider that this ServiceCluster belongs to. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#catalog)

### Tenant.catalog.kubecarrier.io/v1alpha1

Tenant exposes information about available Tenants on the platform and allows a Provider to set custom labels on them.

Tenant objects will be created for all Accounts with the role \"Tenant\" in all Account Namespaces with the role \"Provider\".

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.TenantSpec | false |

[Back to Group](#catalog)

### TenantList.catalog.kubecarrier.io/v1alpha1

TenantList contains a list of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Tenant | true |

[Back to Group](#catalog)

### TenantSpec.catalog.kubecarrier.io/v1alpha1

TenantSpec defines the desired state of Tenant

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#catalog)

### ObjectReference.catalog.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to Group](#catalog)
## Operator

The `operator.kubecarrier.io` API group contains objects to interact with the KubeCarrier installation.

* [Catapult.operator.kubecarrier.io/v1alpha1](#catapult.operator.kubecarrier.io/v1alpha1)
* [CatapultCondition.operator.kubecarrier.io/v1alpha1](#catapultcondition.operator.kubecarrier.io/v1alpha1)
* [CatapultList.operator.kubecarrier.io/v1alpha1](#catapultlist.operator.kubecarrier.io/v1alpha1)
* [CatapultSpec.operator.kubecarrier.io/v1alpha1](#catapultspec.operator.kubecarrier.io/v1alpha1)
* [CatapultStatus.operator.kubecarrier.io/v1alpha1](#catapultstatus.operator.kubecarrier.io/v1alpha1)
* [Elevator.operator.kubecarrier.io/v1alpha1](#elevator.operator.kubecarrier.io/v1alpha1)
* [ElevatorCondition.operator.kubecarrier.io/v1alpha1](#elevatorcondition.operator.kubecarrier.io/v1alpha1)
* [ElevatorList.operator.kubecarrier.io/v1alpha1](#elevatorlist.operator.kubecarrier.io/v1alpha1)
* [ElevatorSpec.operator.kubecarrier.io/v1alpha1](#elevatorspec.operator.kubecarrier.io/v1alpha1)
* [ElevatorStatus.operator.kubecarrier.io/v1alpha1](#elevatorstatus.operator.kubecarrier.io/v1alpha1)
* [Ferry.operator.kubecarrier.io/v1alpha1](#ferry.operator.kubecarrier.io/v1alpha1)
* [FerryCondition.operator.kubecarrier.io/v1alpha1](#ferrycondition.operator.kubecarrier.io/v1alpha1)
* [FerryList.operator.kubecarrier.io/v1alpha1](#ferrylist.operator.kubecarrier.io/v1alpha1)
* [FerrySpec.operator.kubecarrier.io/v1alpha1](#ferryspec.operator.kubecarrier.io/v1alpha1)
* [FerryStatus.operator.kubecarrier.io/v1alpha1](#ferrystatus.operator.kubecarrier.io/v1alpha1)
* [KubeCarrier.operator.kubecarrier.io/v1alpha1](#kubecarrier.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierCondition.operator.kubecarrier.io/v1alpha1](#kubecarriercondition.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierList.operator.kubecarrier.io/v1alpha1](#kubecarrierlist.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierSpec.operator.kubecarrier.io/v1alpha1](#kubecarrierspec.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierStatus.operator.kubecarrier.io/v1alpha1](#kubecarrierstatus.operator.kubecarrier.io/v1alpha1)
* [Tower.operator.kubecarrier.io/v1alpha1](#tower.operator.kubecarrier.io/v1alpha1)
* [TowerCondition.operator.kubecarrier.io/v1alpha1](#towercondition.operator.kubecarrier.io/v1alpha1)
* [TowerList.operator.kubecarrier.io/v1alpha1](#towerlist.operator.kubecarrier.io/v1alpha1)
* [TowerSpec.operator.kubecarrier.io/v1alpha1](#towerspec.operator.kubecarrier.io/v1alpha1)
* [TowerStatus.operator.kubecarrier.io/v1alpha1](#towerstatus.operator.kubecarrier.io/v1alpha1)
* [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreference.operator.kubecarrier.io/v1alpha1)
* [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreference.operator.kubecarrier.io/v1alpha1)

### Catapult.operator.kubecarrier.io/v1alpha1

Catapult manages the deployment of the Catapult controller manager.

A Catapult instance is started for each CustomResourceDiscovery instance and responsible for reconciling CRD instances across Kubernetes Clusters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.CatapultSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.CatapultStatus | false |

[Back to Group](#operator)

### CatapultCondition.operator.kubecarrier.io/v1alpha1

CatapultCondition contains details for the current condition of this Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catapult condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.CatapultConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### CatapultList.operator.kubecarrier.io/v1alpha1

CatapultList contains a list of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Catapult | true |

[Back to Group](#operator)

### CatapultSpec.operator.kubecarrier.io/v1alpha1

CatapultSpec defines the desired state of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| managementClusterCRD | References the CRD in the Management Cluster. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| serviceClusterCRD | References the CRD in the ServiceCluster. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| serviceCluster | References the ServiceCluster object that this object belongs to. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |
| webhookStrategy | WebhookStrategy configs the webhook of the CRD which is registered in the management cluster by this Catapult. There are two possible values for this configuration {None (by default), ServiceCluster} None (by default): Webhook will only check if there is an available ServiceClusterAssignment in the current Namespace. ServiceCluster: Webhook will call webhooks of the CRD in the ServiceCluster with dry-run flag. | corev1alpha1.WebhookStrategyType | false |

[Back to Group](#operator)

### CatapultStatus.operator.kubecarrier.io/v1alpha1

CatapultStatus defines the observed state of Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catapult by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catapult's current state. | []operator.kubecarrier.io/v1alpha1.CatapultCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.CatapultPhaseType | false |

[Back to Group](#operator)

### Elevator.operator.kubecarrier.io/v1alpha1

Elevator manages the deployment of the Elevator controller manager.

For each `DerivedCustomResource` a Elevator instance is launched to propagate the derived CRD instance into the Namespace of it's provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.ElevatorSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.ElevatorStatus | false |

[Back to Group](#operator)

### ElevatorCondition.operator.kubecarrier.io/v1alpha1

ElevatorCondition contains details for the current condition of this Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Elevator condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.ElevatorConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### ElevatorList.operator.kubecarrier.io/v1alpha1

ElevatorList contains a list of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Elevator | true |

[Back to Group](#operator)

### ElevatorSpec.operator.kubecarrier.io/v1alpha1

ElevatorSpec defines the desired state of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| providerCRD | References the provider or internal CRD, that should be created in the provider namespace. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| tenantCRD | References the public CRD that will be synced into the provider namespace. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| derivedCR | References the DerivedCustomResource controlling the Tenant-side CRD. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#operator)

### ElevatorStatus.operator.kubecarrier.io/v1alpha1

ElevatorStatus defines the observed state of Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Elevator by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Elevator's current state. | []operator.kubecarrier.io/v1alpha1.ElevatorCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.ElevatorPhaseType | false |

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
| spec |  | operator.kubecarrier.io/v1alpha1.FerrySpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.FerryStatus | false |

[Back to Group](#operator)

### FerryCondition.operator.kubecarrier.io/v1alpha1

FerryCondition contains details for the current condition of this Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.FerryConditionType | true |

[Back to Group](#operator)

### FerryList.operator.kubecarrier.io/v1alpha1

FerryList contains a list of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Ferry | true |

[Back to Group](#operator)

### FerrySpec.operator.kubecarrier.io/v1alpha1

FerrySpec defines the desired state of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to Group](#operator)

### FerryStatus.operator.kubecarrier.io/v1alpha1

FerryStatus defines the observed state of Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.FerryPhaseType | false |
| conditions | Conditions is a list of all conditions this Ferry is in. | []operator.kubecarrier.io/v1alpha1.FerryCondition | false |
| observedGeneration | The most recent generation observed by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |

[Back to Group](#operator)

### KubeCarrier.operator.kubecarrier.io/v1alpha1

KubeCarrier manages the deployment of the KubeCarrier controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.KubeCarrierSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.KubeCarrierStatus | false |

[Back to Group](#operator)

### KubeCarrierCondition.operator.kubecarrier.io/v1alpha1

KubeCarrierCondition contains details for the current condition of this KubeCarrier.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the KubeCarrier condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.KubeCarrierConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### KubeCarrierList.operator.kubecarrier.io/v1alpha1

KubeCarrierList contains a list of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.KubeCarrier | true |

[Back to Group](#operator)

### KubeCarrierSpec.operator.kubecarrier.io/v1alpha1

KubeCarrierSpec defines the desired state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#operator)

### KubeCarrierStatus.operator.kubecarrier.io/v1alpha1

KubeCarrierStatus defines the observed state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this KubeCarrier by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a KubeCarrier's current state. | []operator.kubecarrier.io/v1alpha1.KubeCarrierCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.KubeCarrierPhaseType | false |

[Back to Group](#operator)

### Tower.operator.kubecarrier.io/v1alpha1

Tower manages the deployment of the KubeCarrier master controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.TowerSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.TowerStatus | false |

[Back to Group](#operator)

### TowerCondition.operator.kubecarrier.io/v1alpha1

TowerCondition contains details for the current condition of this Tower.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Tower condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.TowerConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to Group](#operator)

### TowerList.operator.kubecarrier.io/v1alpha1

TowerList contains a list of Tower

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Tower | true |

[Back to Group](#operator)

### TowerSpec.operator.kubecarrier.io/v1alpha1

TowerSpec defines the desired state of Tower

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to Group](#operator)

### TowerStatus.operator.kubecarrier.io/v1alpha1

TowerStatus defines the observed state of Tower

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Tower by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Tower's current state. | []operator.kubecarrier.io/v1alpha1.TowerCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.TowerPhaseType | false |

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
