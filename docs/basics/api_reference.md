
## Table of Contents
* [CustomResourceDiscoverySet.kubecarrier.io/v1alpha1](#customresourcediscoveryset.kubecarrier.io/v1alpha1)
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
* [ObjectReference.kubecarrier.io/v1alpha1](#objectreference.kubecarrier.io/v1alpha1)
* [CustomResourceDiscovery.kubecarrier.io/v1alpha1](#customresourcediscovery.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1](#customresourcediscoverycondition.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryList.kubecarrier.io/v1alpha1](#customresourcediscoverylist.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1](#customresourcediscoveryspec.kubecarrier.io/v1alpha1)
* [CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1](#customresourcediscoverystatus.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignment.kubecarrier.io/v1alpha1](#serviceclusterassignment.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1](#serviceclusterassignmentcondition.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentList.kubecarrier.io/v1alpha1](#serviceclusterassignmentlist.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1](#serviceclusterassignmentspec.kubecarrier.io/v1alpha1)
* [ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1](#serviceclusterassignmentstatus.kubecarrier.io/v1alpha1)
* [TenantReference.catalog.kubecarrier.io/v1alpha1](#tenantreference.catalog.kubecarrier.io/v1alpha1)
* [TenantReferenceList.catalog.kubecarrier.io/v1alpha1](#tenantreferencelist.catalog.kubecarrier.io/v1alpha1)
* [TenantReferenceSpec.catalog.kubecarrier.io/v1alpha1](#tenantreferencespec.catalog.kubecarrier.io/v1alpha1)
* [CRDInformation.catalog.kubecarrier.io/v1alpha1](#crdinformation.catalog.kubecarrier.io/v1alpha1)
* [CRDVersion.catalog.kubecarrier.io/v1alpha1](#crdversion.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResource.catalog.kubecarrier.io/v1alpha1](#derivedcustomresource.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcecondition.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcelist.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceReference.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcereference.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcespec.catalog.kubecarrier.io/v1alpha1)
* [DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1](#derivedcustomresourcestatus.catalog.kubecarrier.io/v1alpha1)
* [FieldPath.catalog.kubecarrier.io/v1alpha1](#fieldpath.catalog.kubecarrier.io/v1alpha1)
* [VersionExposeConfig.catalog.kubecarrier.io/v1alpha1](#versionexposeconfig.catalog.kubecarrier.io/v1alpha1)
* [ProviderReference.catalog.kubecarrier.io/v1alpha1](#providerreference.catalog.kubecarrier.io/v1alpha1)
* [ProviderReferenceList.catalog.kubecarrier.io/v1alpha1](#providerreferencelist.catalog.kubecarrier.io/v1alpha1)
* [ProviderReferenceSpec.catalog.kubecarrier.io/v1alpha1](#providerreferencespec.catalog.kubecarrier.io/v1alpha1)
* [Catalog.catalog.kubecarrier.io/v1alpha1](#catalog.catalog.kubecarrier.io/v1alpha1)
* [CatalogCondition.catalog.kubecarrier.io/v1alpha1](#catalogcondition.catalog.kubecarrier.io/v1alpha1)
* [CatalogList.catalog.kubecarrier.io/v1alpha1](#cataloglist.catalog.kubecarrier.io/v1alpha1)
* [CatalogSpec.catalog.kubecarrier.io/v1alpha1](#catalogspec.catalog.kubecarrier.io/v1alpha1)
* [CatalogStatus.catalog.kubecarrier.io/v1alpha1](#catalogstatus.catalog.kubecarrier.io/v1alpha1)
* [ObjectReference.catalog.kubecarrier.io/v1alpha1](#objectreference.catalog.kubecarrier.io/v1alpha1)
* [Provider.catalog.kubecarrier.io/v1alpha1](#provider.catalog.kubecarrier.io/v1alpha1)
* [ProviderCondition.catalog.kubecarrier.io/v1alpha1](#providercondition.catalog.kubecarrier.io/v1alpha1)
* [ProviderList.catalog.kubecarrier.io/v1alpha1](#providerlist.catalog.kubecarrier.io/v1alpha1)
* [ProviderMetadata.catalog.kubecarrier.io/v1alpha1](#providermetadata.catalog.kubecarrier.io/v1alpha1)
* [ProviderSpec.catalog.kubecarrier.io/v1alpha1](#providerspec.catalog.kubecarrier.io/v1alpha1)
* [ProviderStatus.catalog.kubecarrier.io/v1alpha1](#providerstatus.catalog.kubecarrier.io/v1alpha1)
* [ServiceClusterReference.catalog.kubecarrier.io/v1alpha1](#serviceclusterreference.catalog.kubecarrier.io/v1alpha1)
* [ServiceClusterReferenceList.catalog.kubecarrier.io/v1alpha1](#serviceclusterreferencelist.catalog.kubecarrier.io/v1alpha1)
* [ServiceClusterReferenceSpec.catalog.kubecarrier.io/v1alpha1](#serviceclusterreferencespec.catalog.kubecarrier.io/v1alpha1)
* [Tenant.catalog.kubecarrier.io/v1alpha1](#tenant.catalog.kubecarrier.io/v1alpha1)
* [TenantCondition.catalog.kubecarrier.io/v1alpha1](#tenantcondition.catalog.kubecarrier.io/v1alpha1)
* [TenantList.catalog.kubecarrier.io/v1alpha1](#tenantlist.catalog.kubecarrier.io/v1alpha1)
* [TenantSpec.catalog.kubecarrier.io/v1alpha1](#tenantspec.catalog.kubecarrier.io/v1alpha1)
* [TenantStatus.catalog.kubecarrier.io/v1alpha1](#tenantstatus.catalog.kubecarrier.io/v1alpha1)
* [Offering.catalog.kubecarrier.io/v1alpha1](#offering.catalog.kubecarrier.io/v1alpha1)
* [OfferingData.catalog.kubecarrier.io/v1alpha1](#offeringdata.catalog.kubecarrier.io/v1alpha1)
* [OfferingList.catalog.kubecarrier.io/v1alpha1](#offeringlist.catalog.kubecarrier.io/v1alpha1)
* [OfferingMetadata.catalog.kubecarrier.io/v1alpha1](#offeringmetadata.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntry.catalog.kubecarrier.io/v1alpha1](#catalogentry.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1](#catalogentrycondition.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryList.catalog.kubecarrier.io/v1alpha1](#catalogentrylist.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1](#catalogentrymetadata.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1](#catalogentryspec.catalog.kubecarrier.io/v1alpha1)
* [CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1](#catalogentrystatus.catalog.kubecarrier.io/v1alpha1)
* [DerivedConfig.catalog.kubecarrier.io/v1alpha1](#derivedconfig.catalog.kubecarrier.io/v1alpha1)
* [KubeCarrier.operator.kubecarrier.io/v1alpha1](#kubecarrier.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierCondition.operator.kubecarrier.io/v1alpha1](#kubecarriercondition.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierList.operator.kubecarrier.io/v1alpha1](#kubecarrierlist.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierSpec.operator.kubecarrier.io/v1alpha1](#kubecarrierspec.operator.kubecarrier.io/v1alpha1)
* [KubeCarrierStatus.operator.kubecarrier.io/v1alpha1](#kubecarrierstatus.operator.kubecarrier.io/v1alpha1)
* [CRDReference.operator.kubecarrier.io/v1alpha1](#crdreference.operator.kubecarrier.io/v1alpha1)
* [ObjectReference.operator.kubecarrier.io/v1alpha1](#objectreference.operator.kubecarrier.io/v1alpha1)
* [Ferry.operator.kubecarrier.io/v1alpha1](#ferry.operator.kubecarrier.io/v1alpha1)
* [FerryCondition.operator.kubecarrier.io/v1alpha1](#ferrycondition.operator.kubecarrier.io/v1alpha1)
* [FerryList.operator.kubecarrier.io/v1alpha1](#ferrylist.operator.kubecarrier.io/v1alpha1)
* [FerrySpec.operator.kubecarrier.io/v1alpha1](#ferryspec.operator.kubecarrier.io/v1alpha1)
* [FerryStatus.operator.kubecarrier.io/v1alpha1](#ferrystatus.operator.kubecarrier.io/v1alpha1)
* [Elevator.operator.kubecarrier.io/v1alpha1](#elevator.operator.kubecarrier.io/v1alpha1)
* [ElevatorCondition.operator.kubecarrier.io/v1alpha1](#elevatorcondition.operator.kubecarrier.io/v1alpha1)
* [ElevatorList.operator.kubecarrier.io/v1alpha1](#elevatorlist.operator.kubecarrier.io/v1alpha1)
* [ElevatorSpec.operator.kubecarrier.io/v1alpha1](#elevatorspec.operator.kubecarrier.io/v1alpha1)
* [ElevatorStatus.operator.kubecarrier.io/v1alpha1](#elevatorstatus.operator.kubecarrier.io/v1alpha1)
* [Catapult.operator.kubecarrier.io/v1alpha1](#catapult.operator.kubecarrier.io/v1alpha1)
* [CatapultCondition.operator.kubecarrier.io/v1alpha1](#catapultcondition.operator.kubecarrier.io/v1alpha1)
* [CatapultList.operator.kubecarrier.io/v1alpha1](#catapultlist.operator.kubecarrier.io/v1alpha1)
* [CatapultSpec.operator.kubecarrier.io/v1alpha1](#catapultspec.operator.kubecarrier.io/v1alpha1)
* [CatapultStatus.operator.kubecarrier.io/v1alpha1](#catapultstatus.operator.kubecarrier.io/v1alpha1)

## CustomResourceDiscoverySet.kubecarrier.io/v1alpha1

CustomResourceDiscoverySet manages multiple CustomResourceDiscovery objects for a set of service clusters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetSpec | false |
| status |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetStatus | false |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoverySetCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoverySetCondition contains details for the current condition of this CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetConditionType | true |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoverySetList.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.CustomResourceDiscoverySet | true |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoverySetSpec.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| serviceClusterSelector | ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | true |
| kindOverride | KindOverride overrides resulting internal CRDs kind | string | false |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoverySetStatus.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.CustomResourceDiscoverySetPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | []kubecarrier.io/v1alpha1.CustomResourceDiscoverySetCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## ServiceCluster.kubecarrier.io/v1alpha1

ServiceCluster is a providers Kubernetes Cluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | kubecarrier.io/v1alpha1.ServiceClusterSpec | false |
| status |  | kubecarrier.io/v1alpha1.ServiceClusterStatus | false |

[Back to TOC](#table-of-contents)

## ServiceClusterCondition.kubecarrier.io/v1alpha1

ServiceClusterCondition contains details for the current condition of this ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastHeartbeatTime | LastHeartbeatTime is the timestamp corresponding to the last update of this condition. | metav1.Time | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.ServiceClusterConditionType | true |

[Back to TOC](#table-of-contents)

## ServiceClusterList.kubecarrier.io/v1alpha1

ServiceClusterList contains a list of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.ServiceCluster | true |

[Back to TOC](#table-of-contents)

## ServiceClusterMetadata.kubecarrier.io/v1alpha1

ServiceClusterMetadata contains the metadata (display name, description, etc) of the ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this ServiceCluster. | string | false |
| description | Description shows the human-readable description of this ServiceCluster. | string | false |

[Back to TOC](#table-of-contents)

## ServiceClusterSpec.kubecarrier.io/v1alpha1

ServiceClusterSpec defines the desired state of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | kubecarrier.io/v1alpha1.ServiceClusterMetadata | false |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## ServiceClusterStatus.kubecarrier.io/v1alpha1

ServiceClusterStatus defines the observed state of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.ServiceClusterPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceCluster is in. | []kubecarrier.io/v1alpha1.ServiceClusterCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |
| kubernetesVersion | KubernetesVersion of the service cluster API Server | *version.Info | false |

[Back to TOC](#table-of-contents)

## ObjectReference.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)

## CustomResourceDiscovery.kubecarrier.io/v1alpha1

CustomResourceDiscovery is used inside KubeCarrier to fetch a CustomResourceDefinition from another cluster and to offload cross cluster access to another component.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | kubecarrier.io/v1alpha1.CustomResourceDiscoverySpec | false |
| status |  | kubecarrier.io/v1alpha1.CustomResourceDiscoveryStatus | false |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoveryCondition.kubecarrier.io/v1alpha1

CustomResourceDiscoveryCondition contains details for the current condition of this CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.CustomResourceDiscoveryConditionType | true |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoveryList.kubecarrier.io/v1alpha1

CustomResourceDiscoveryList contains a list of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.CustomResourceDiscovery | true |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoverySpec.kubecarrier.io/v1alpha1

CustomResourceDiscoverySpec defines the desired state of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| serviceCluster | ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on. | kubecarrier.io/v1alpha1.ObjectReference | true |
| kindOverride | KindOverride overrides resulting internal CRDs kind | string | false |

[Back to TOC](#table-of-contents)

## CustomResourceDiscoveryStatus.kubecarrier.io/v1alpha1

CustomResourceDiscoveryStatus defines the observed state of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD defines the original CustomResourceDefinition specification from the service cluster | *apiextensionsv1.CustomResourceDefinition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.CustomResourceDiscoveryPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | []kubecarrier.io/v1alpha1.CustomResourceDiscoveryCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## ServiceClusterAssignment.kubecarrier.io/v1alpha1

ServiceClusterAssignment represents the assignment of a Tenant to a ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | kubecarrier.io/v1alpha1.ServiceClusterAssignmentSpec | false |
| status |  | kubecarrier.io/v1alpha1.ServiceClusterAssignmentStatus | false |

[Back to TOC](#table-of-contents)

## ServiceClusterAssignmentCondition.kubecarrier.io/v1alpha1

ServiceClusterAssignmentCondition contains details for the current condition of this ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | kubecarrier.io/v1alpha1.ServiceClusterAssignmentConditionType | true |

[Back to TOC](#table-of-contents)

## ServiceClusterAssignmentList.kubecarrier.io/v1alpha1

ServiceClusterAssignmentList contains a list of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []kubecarrier.io/v1alpha1.ServiceClusterAssignment | true |

[Back to TOC](#table-of-contents)

## ServiceClusterAssignmentSpec.kubecarrier.io/v1alpha1

ServiceClusterAssignmentSpec defines the desired state of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| serviceCluster | References the ServiceCluster. | kubecarrier.io/v1alpha1.ObjectReference | true |
| managementNamespace | References the source namespace in the management cluster. | kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## ServiceClusterAssignmentStatus.kubecarrier.io/v1alpha1

ServiceClusterAssignmentStatus defines the observed state of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | kubecarrier.io/v1alpha1.ServiceClusterAssignmentPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceClusterAssignment is in. | []kubecarrier.io/v1alpha1.ServiceClusterAssignmentCondition | false |
| observedGeneration | The most recent generation observed by the controller. | kubecarrier.io/v1alpha1.int64 | false |
| serviceClusterNamespace | ServiceClusterNamespace references the Namespace on the ServiceCluster that was assigned. | kubecarrier.io/v1alpha1.ObjectReference | false |

[Back to TOC](#table-of-contents)

## TenantReference.catalog.kubecarrier.io/v1alpha1

TenantReference is a read-only object exposing the Tenant information. TenantReference lives in the provider's namespace. The provider is allowed modifying TenantReference's labels, marking them at will. This allows the tenant granular tenant selection for the offered services catalogs.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.TenantReferenceSpec | false |

[Back to TOC](#table-of-contents)

## TenantReferenceList.catalog.kubecarrier.io/v1alpha1

TenantReferenceList contains a list of TenantReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.TenantReference | true |

[Back to TOC](#table-of-contents)

## TenantReferenceSpec.catalog.kubecarrier.io/v1alpha1

TenantReferenceSpec defines the desired state of TenantReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## CRDInformation.catalog.kubecarrier.io/v1alpha1

CRDInformation contains type information about the CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| apiGroup |  | string | true |
| kind |  | string | true |
| versions |  | []catalog.kubecarrier.io/v1alpha1.CRDVersion | true |
| serviceCluster | ServiceCluster references a ServiceClusterReference of this CRD. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## CRDVersion.catalog.kubecarrier.io/v1alpha1

CRDVersion holds CRD version specific details.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of this version, for example: v1, v1alpha1, v1beta1 | string | true |
| schema | Schema of this CRD version. | *apiextensionsv1.CustomResourceValidation | false |

[Back to TOC](#table-of-contents)

## DerivedCustomResource.catalog.kubecarrier.io/v1alpha1

DerivedCustomResource derives a new CRD from a existing one.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceStatus | false |

[Back to TOC](#table-of-contents)

## DerivedCustomResourceCondition.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceCondition contains details for the current condition of this DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the DerivedCustomResource condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## DerivedCustomResourceList.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceList contains a list of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.DerivedCustomResource | true |

[Back to TOC](#table-of-contents)

## DerivedCustomResourceReference.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceReference references the derived CRD controlled by this DerivedCustomResource instance.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the derived CRD. | string | true |
| group | API Group of the derived CRD. | string | true |
| kind |  | string | true |
| plural |  | string | true |
| singular |  | string | true |

[Back to TOC](#table-of-contents)

## DerivedCustomResourceSpec.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceSpec defines the desired state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| baseCRD | CRD that should be used as a base to derive a new CRD from. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.VersionExposeConfig | true |

[Back to TOC](#table-of-contents)

## DerivedCustomResourceStatus.catalog.kubecarrier.io/v1alpha1

DerivedCustomResourceStatus defines the observed state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this DerivedCustomResource by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a DerivedCustomResource's current state. | []catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.DerivedCustomResourcePhaseType | false |
| derivedCR | DerivedCR holds information about the derived CRD. | *catalog.kubecarrier.io/v1alpha1.DerivedCustomResourceReference | false |

[Back to TOC](#table-of-contents)

## FieldPath.catalog.kubecarrier.io/v1alpha1

FieldPath is specifying how to address a certain field.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jsonPath | JSONPath e.g. .spec.somefield.somesubfield | string | true |

[Back to TOC](#table-of-contents)

## VersionExposeConfig.catalog.kubecarrier.io/v1alpha1

VersionExposeConfig specifies which fields to expose in the derived CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| versions | specifies the versions of the referenced CRD, that this expose config applies to. The same version may not be specified in multiple VersionExposeConfigs. | []string | true |
| fields | specifies the fields that should be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.FieldPath | true |

[Back to TOC](#table-of-contents)

## ProviderReference.catalog.kubecarrier.io/v1alpha1

ProviderReference exposes information of the Provider(displayName, description). This object lives in the tenant namespace for each provider the tenant is allowed utilizing (e.g. there's catalog selecting this tenant as its user)

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.ProviderReferenceSpec | false |

[Back to TOC](#table-of-contents)

## ProviderReferenceList.catalog.kubecarrier.io/v1alpha1

ProviderReferenceList contains a list of ProviderReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.ProviderReference | true |

[Back to TOC](#table-of-contents)

## ProviderReferenceSpec.catalog.kubecarrier.io/v1alpha1

ProviderReferenceSpec defines the desired state of ProviderReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the Provider. | catalog.kubecarrier.io/v1alpha1.ProviderMetadata | false |

[Back to TOC](#table-of-contents)

## Catalog.catalog.kubecarrier.io/v1alpha1

Catalog publishes a selection of CatalogEntries to a selection of Tenants.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.CatalogSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.CatalogStatus | false |

[Back to TOC](#table-of-contents)

## CatalogCondition.catalog.kubecarrier.io/v1alpha1

CatalogCondition contains details for the current condition of this Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catalog condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.CatalogConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## CatalogList.catalog.kubecarrier.io/v1alpha1

CatalogList contains a list of Catalog

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Catalog | true |

[Back to TOC](#table-of-contents)

## CatalogSpec.catalog.kubecarrier.io/v1alpha1

CatalogSpec defines the desired contents of a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| catalogEntrySelector | CatalogEntrySelector selects CatalogEntry objects that should be part of this catalog. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |
| tenantReferenceSelector | TenantReferenceSelector selects TenantReference objects that the catalog should be published to. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |

[Back to TOC](#table-of-contents)

## CatalogStatus.catalog.kubecarrier.io/v1alpha1

CatalogStatus defines the observed state of Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tenants | Tenants is the list of the Tenants(TenantReference) that selected by this Catalog. | []catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| entries | Entries is the list of the CatalogEntries that selected by this Catalog. | []catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catalog by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catalog's current state. | []catalog.kubecarrier.io/v1alpha1.CatalogCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.CatalogPhaseType | false |

[Back to TOC](#table-of-contents)

## ObjectReference.catalog.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)

## Provider.catalog.kubecarrier.io/v1alpha1

Provider is the service provider representation in the KubeCarrier control-plane.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.ProviderSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.ProviderStatus | false |

[Back to TOC](#table-of-contents)

## ProviderCondition.catalog.kubecarrier.io/v1alpha1

ProviderCondition contains details for the current condition of this Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Provider condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.ProviderConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## ProviderList.catalog.kubecarrier.io/v1alpha1

ProviderList contains a list of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Provider | true |

[Back to TOC](#table-of-contents)

## ProviderMetadata.catalog.kubecarrier.io/v1alpha1

ProviderMetadata contains the metadata (display name, description, etc) of the Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this Provider. | string | false |
| description | Description shows the human-readable description of this Provider. | string | false |

[Back to TOC](#table-of-contents)

## ProviderSpec.catalog.kubecarrier.io/v1alpha1

ProviderSpec defines the desired state of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | catalog.kubecarrier.io/v1alpha1.ProviderMetadata | false |

[Back to TOC](#table-of-contents)

## ProviderStatus.catalog.kubecarrier.io/v1alpha1

ProviderStatus defines the observed state of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespaceName | NamespaceName is the name of the namespace that the Provider manages. | string | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Provider by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Provider's current state. | []catalog.kubecarrier.io/v1alpha1.ProviderCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.ProviderPhaseType | false |

[Back to TOC](#table-of-contents)

## ServiceClusterReference.catalog.kubecarrier.io/v1alpha1

ServiceClusterReference exposes information about a Providers Clusters.\n\nThis object lives in the tenant namespace for each provider the tenant is allowed utilizing (e.g. there's catalog selecting this tenant as its user)

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.ServiceClusterReferenceSpec | false |

[Back to TOC](#table-of-contents)

## ServiceClusterReferenceList.catalog.kubecarrier.io/v1alpha1

ServiceClusterReferenceList contains a list of ServiceClusterReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.ServiceClusterReference | true |

[Back to TOC](#table-of-contents)

## ServiceClusterReferenceSpec.catalog.kubecarrier.io/v1alpha1

ServiceClusterReferenceSpec defines the desired state of ServiceClusterReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the ServiceCluster. | corev1alpha1.ServiceClusterMetadata | false |
| provider | Provider references the Provider that this ServiceCluster belongs to. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## Tenant.catalog.kubecarrier.io/v1alpha1

Tenant sets up permissions and references to allow a end-user group to interact with providers' services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.TenantSpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.TenantStatus | false |

[Back to TOC](#table-of-contents)

## TenantCondition.catalog.kubecarrier.io/v1alpha1

TenantCondition contains details for the current condition of this Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Tenant condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.TenantConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## TenantList.catalog.kubecarrier.io/v1alpha1

TenantList contains a list of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Tenant | true |

[Back to TOC](#table-of-contents)

## TenantSpec.catalog.kubecarrier.io/v1alpha1

TenantSpec defines the desired state of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## TenantStatus.catalog.kubecarrier.io/v1alpha1

TenantStatus defines the observed state of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespaceName | NamespaceName is the name of the namespace that the Tenant manages. | string | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Tenant by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Tenant's current state. | []catalog.kubecarrier.io/v1alpha1.TenantCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.TenantPhaseType | false |

[Back to TOC](#table-of-contents)

## Offering.catalog.kubecarrier.io/v1alpha1

Offering is used for Tenants to discover services that have been made available to them.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| offering |  | catalog.kubecarrier.io/v1alpha1.OfferingData | false |

[Back to TOC](#table-of-contents)

## OfferingData.catalog.kubecarrier.io/v1alpha1

OfferingData defines the data (metadata, provider, crds, etc.) of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | catalog.kubecarrier.io/v1alpha1.OfferingMetadata | false |
| provider | Provider references a ProviderReference of this Offering. | catalog.kubecarrier.io/v1alpha1.ObjectReference | true |
| crd | CRD holds the information about the underlying CRD that is offered by this offering. | catalog.kubecarrier.io/v1alpha1.CRDInformation | false |

[Back to TOC](#table-of-contents)

## OfferingList.catalog.kubecarrier.io/v1alpha1

OfferingList contains a list of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.Offering | true |

[Back to TOC](#table-of-contents)

## OfferingMetadata.catalog.kubecarrier.io/v1alpha1

OfferingMetadata contains the metadata (display name, description, etc) of the Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this Offering. | string | false |
| description | Description shows the human-readable description of this Offering. | string | false |

[Back to TOC](#table-of-contents)

## CatalogEntry.catalog.kubecarrier.io/v1alpha1

CatalogEntry reference to the CRD that the provider wants to provide as service to the tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | catalog.kubecarrier.io/v1alpha1.CatalogEntrySpec | false |
| status |  | catalog.kubecarrier.io/v1alpha1.CatalogEntryStatus | false |

[Back to TOC](#table-of-contents)

## CatalogEntryCondition.catalog.kubecarrier.io/v1alpha1

CatalogEntryCondition contains details for the current condition of this CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntry condition, currently ('Ready'). | catalog.kubecarrier.io/v1alpha1.CatalogEntryConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalog.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## CatalogEntryList.catalog.kubecarrier.io/v1alpha1

CatalogEntryList contains a list of CatalogEntry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []catalog.kubecarrier.io/v1alpha1.CatalogEntry | true |

[Back to TOC](#table-of-contents)

## CatalogEntryMetadata.catalog.kubecarrier.io/v1alpha1

CatalogEntryMetadata contains the metadata (display name, description, etc) of the CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this CatalogEntry. | string | false |
| description | Description shows the human-readable description of this CatalogEntry. | string | false |

[Back to TOC](#table-of-contents)

## CatalogEntrySpec.catalog.kubecarrier.io/v1alpha1

CatalogEntrySpec defines the desired state of CatalogEntry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the CatalogEntry. | catalog.kubecarrier.io/v1alpha1.CatalogEntryMetadata | false |
| baseCRD | BaseCRD is the underlying BaseCRD objects that this CatalogEntry refers to. | catalog.kubecarrier.io/v1alpha1.ObjectReference | false |
| derivedConfig | DerivedConfig contains the configuration to generate DerivedCustomResource from the BaseCRD of this CatalogEntry. | *catalog.kubecarrier.io/v1alpha1.DerivedConfig | false |

[Back to TOC](#table-of-contents)

## CatalogEntryStatus.catalog.kubecarrier.io/v1alpha1

CatalogEntryStatus defines the observed state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD holds the information about the underlying BaseCRD that are offered by this CatalogEntry. | *catalog.kubecarrier.io/v1alpha1.CRDInformation | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntry by the controller. | catalog.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntry's current state. | []catalog.kubecarrier.io/v1alpha1.CatalogEntryCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalog.kubecarrier.io/v1alpha1.CatalogEntryPhaseType | false |

[Back to TOC](#table-of-contents)

## DerivedConfig.catalog.kubecarrier.io/v1alpha1



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | []catalog.kubecarrier.io/v1alpha1.VersionExposeConfig | true |

[Back to TOC](#table-of-contents)

## KubeCarrier.operator.kubecarrier.io/v1alpha1

KubeCarrier manages the deployment of the KubeCarrier controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.KubeCarrierSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.KubeCarrierStatus | false |

[Back to TOC](#table-of-contents)

## KubeCarrierCondition.operator.kubecarrier.io/v1alpha1

KubeCarrierCondition contains details for the current condition of this KubeCarrier.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the KubeCarrier condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.KubeCarrierConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## KubeCarrierList.operator.kubecarrier.io/v1alpha1

KubeCarrierList contains a list of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.KubeCarrier | true |

[Back to TOC](#table-of-contents)

## KubeCarrierSpec.operator.kubecarrier.io/v1alpha1

KubeCarrierSpec defines the desired state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## KubeCarrierStatus.operator.kubecarrier.io/v1alpha1

KubeCarrierStatus defines the observed state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this KubeCarrier by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a KubeCarrier's current state. | []operator.kubecarrier.io/v1alpha1.KubeCarrierCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.KubeCarrierPhaseType | false |

[Back to TOC](#table-of-contents)

## CRDReference.operator.kubecarrier.io/v1alpha1

CRDReference references a CustomResourceDefitition.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kind |  | string | true |
| version |  | string | true |
| group |  | string | true |
| plural |  | string | true |

[Back to TOC](#table-of-contents)

## ObjectReference.operator.kubecarrier.io/v1alpha1

ObjectReference describes the link to another object in the same namespace

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)

## Ferry.operator.kubecarrier.io/v1alpha1

Ferry represents single kubernetes cluster belonging to the provider\n\nFerry lives in the provider namespace. For each ferry the kubecarrier operator spins up the ferry controller deployment, necessary roles, service accounts, and role bindings\n\nThe reason for ferry controller deployment are multiples: * security --> kubecarrier operator has greater privileges then ferry controller * resource isolation --> each ferry controller pod operates only on a single service cluster,\n\t\tthus resource allocation and monitoring is separate per ferry. This allows finer grade\n\t\tresource tuning and monitoring\n* flexibility --> If needed different ferries could have different deployments depending on\n\t\ttheir specific need (e.g. kubecarrier image version for gradual rolling upgrade, different resource allocation, etc),

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.FerrySpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.FerryStatus | false |

[Back to TOC](#table-of-contents)

## FerryCondition.operator.kubecarrier.io/v1alpha1

FerryCondition contains details for the current condition of this Ferry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.FerryConditionType | true |

[Back to TOC](#table-of-contents)

## FerryList.operator.kubecarrier.io/v1alpha1

FerryList contains a list of Ferry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Ferry | true |

[Back to TOC](#table-of-contents)

## FerrySpec.operator.kubecarrier.io/v1alpha1

FerrySpec defines the desired state of Ferry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## FerryStatus.operator.kubecarrier.io/v1alpha1

FerryStatus defines the observed state of Ferry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.FerryPhaseType | false |
| conditions | Conditions is a list of all conditions this Ferry is in. | []operator.kubecarrier.io/v1alpha1.FerryCondition | false |
| observedGeneration | The most recent generation observed by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## Elevator.operator.kubecarrier.io/v1alpha1

Elevator manages the deployment of the Elevator controller manager. For each `DerivedCustomResource` a Elevator instance is launched to propagate the derived CRD instance into the provider namespace. This component works hand-in-hand with the Catapult instance for the respective type.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.ElevatorSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.ElevatorStatus | false |

[Back to TOC](#table-of-contents)

## ElevatorCondition.operator.kubecarrier.io/v1alpha1

ElevatorCondition contains details for the current condition of this Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Elevator condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.ElevatorConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## ElevatorList.operator.kubecarrier.io/v1alpha1

ElevatorList contains a list of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Elevator | true |

[Back to TOC](#table-of-contents)

## ElevatorSpec.operator.kubecarrier.io/v1alpha1

ElevatorSpec defines the desired state of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| providerCRD | References the provider or internal CRD, that should be created in the provider namespace. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| tenantCRD | References the public CRD that will be synced into the provider namespace. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| derivedCR | References the DerivedCustomResource controlling the Tenant-side CRD. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## ElevatorStatus.operator.kubecarrier.io/v1alpha1

ElevatorStatus defines the observed state of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Elevator by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Elevator's current state. | []operator.kubecarrier.io/v1alpha1.ElevatorCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.ElevatorPhaseType | false |

[Back to TOC](#table-of-contents)

## Catapult.operator.kubecarrier.io/v1alpha1

Catapult manages the deployment of the Catapult controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | operator.kubecarrier.io/v1alpha1.CatapultSpec | false |
| status |  | operator.kubecarrier.io/v1alpha1.CatapultStatus | false |

[Back to TOC](#table-of-contents)

## CatapultCondition.operator.kubecarrier.io/v1alpha1

CatapultCondition contains details for the current condition of this Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catapult condition, currently ('Ready'). | operator.kubecarrier.io/v1alpha1.CatapultConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operator.kubecarrier.io/v1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## CatapultList.operator.kubecarrier.io/v1alpha1

CatapultList contains a list of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | []operator.kubecarrier.io/v1alpha1.Catapult | true |

[Back to TOC](#table-of-contents)

## CatapultSpec.operator.kubecarrier.io/v1alpha1

CatapultSpec defines the desired state of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| managementClusterCRD | References the CRD in the management cluster. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| serviceClusterCRD | References the CRD in the service cluster. | operator.kubecarrier.io/v1alpha1.CRDReference | true |
| serviceCluster | References the ServiceCluster object that this object belongs to. | operator.kubecarrier.io/v1alpha1.ObjectReference | true |

[Back to TOC](#table-of-contents)

## CatapultStatus.operator.kubecarrier.io/v1alpha1

CatapultStatus defines the observed state of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catapult by the controller. | operator.kubecarrier.io/v1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catapult's current state. | []operator.kubecarrier.io/v1alpha1.CatapultCondition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operator.kubecarrier.io/v1alpha1.CatapultPhaseType | false |

[Back to TOC](#table-of-contents)
