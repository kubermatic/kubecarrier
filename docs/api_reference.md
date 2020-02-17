---
title: "API reference"
date: 2020-02-14
weight: 1500
---


## Table of Contents
* [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference)
* [catalogv1alpha1.Catalog](#catalogv1alpha1.catalog)
* [catalogv1alpha1.CatalogCondition](#catalogv1alpha1.catalogcondition)
* [catalogv1alpha1.CatalogList](#catalogv1alpha1.cataloglist)
* [catalogv1alpha1.CatalogSpec](#catalogv1alpha1.catalogspec)
* [catalogv1alpha1.CatalogStatus](#catalogv1alpha1.catalogstatus)
* [catalogv1alpha1.CatalogEntry](#catalogv1alpha1.catalogentry)
* [catalogv1alpha1.CatalogEntryCondition](#catalogv1alpha1.catalogentrycondition)
* [catalogv1alpha1.CatalogEntryList](#catalogv1alpha1.catalogentrylist)
* [catalogv1alpha1.CatalogEntryMetadata](#catalogv1alpha1.catalogentrymetadata)
* [catalogv1alpha1.CatalogEntrySpec](#catalogv1alpha1.catalogentryspec)
* [catalogv1alpha1.CatalogEntryStatus](#catalogv1alpha1.catalogentrystatus)
* [catalogv1alpha1.CRDInformation](#catalogv1alpha1.crdinformation)
* [catalogv1alpha1.CRDVersion](#catalogv1alpha1.crdversion)
* [catalogv1alpha1.DerivedCustomResource](#catalogv1alpha1.derivedcustomresource)
* [catalogv1alpha1.DerivedCustomResourceCondition](#catalogv1alpha1.derivedcustomresourcecondition)
* [catalogv1alpha1.DerivedCustomResourceList](#catalogv1alpha1.derivedcustomresourcelist)
* [catalogv1alpha1.DerivedCustomResourceReference](#catalogv1alpha1.derivedcustomresourcereference)
* [catalogv1alpha1.DerivedCustomResourceSpec](#catalogv1alpha1.derivedcustomresourcespec)
* [catalogv1alpha1.DerivedCustomResourceStatus](#catalogv1alpha1.derivedcustomresourcestatus)
* [catalogv1alpha1.FieldPath](#catalogv1alpha1.fieldpath)
* [catalogv1alpha1.VersionExposeConfig](#catalogv1alpha1.versionexposeconfig)
* [catalogv1alpha1.Offering](#catalogv1alpha1.offering)
* [catalogv1alpha1.OfferingData](#catalogv1alpha1.offeringdata)
* [catalogv1alpha1.OfferingList](#catalogv1alpha1.offeringlist)
* [catalogv1alpha1.OfferingMetadata](#catalogv1alpha1.offeringmetadata)
* [catalogv1alpha1.Provider](#catalogv1alpha1.provider)
* [catalogv1alpha1.ProviderCondition](#catalogv1alpha1.providercondition)
* [catalogv1alpha1.ProviderList](#catalogv1alpha1.providerlist)
* [catalogv1alpha1.ProviderMetadata](#catalogv1alpha1.providermetadata)
* [catalogv1alpha1.ProviderSpec](#catalogv1alpha1.providerspec)
* [catalogv1alpha1.ProviderStatus](#catalogv1alpha1.providerstatus)
* [catalogv1alpha1.ProviderReference](#catalogv1alpha1.providerreference)
* [catalogv1alpha1.ProviderReferenceList](#catalogv1alpha1.providerreferencelist)
* [catalogv1alpha1.ProviderReferenceSpec](#catalogv1alpha1.providerreferencespec)
* [catalogv1alpha1.ServiceClusterReference](#catalogv1alpha1.serviceclusterreference)
* [catalogv1alpha1.ServiceClusterReferenceList](#catalogv1alpha1.serviceclusterreferencelist)
* [catalogv1alpha1.ServiceClusterReferenceSpec](#catalogv1alpha1.serviceclusterreferencespec)
* [catalogv1alpha1.Tenant](#catalogv1alpha1.tenant)
* [catalogv1alpha1.TenantCondition](#catalogv1alpha1.tenantcondition)
* [catalogv1alpha1.TenantList](#catalogv1alpha1.tenantlist)
* [catalogv1alpha1.TenantSpec](#catalogv1alpha1.tenantspec)
* [catalogv1alpha1.TenantStatus](#catalogv1alpha1.tenantstatus)
* [catalogv1alpha1.TenantReference](#catalogv1alpha1.tenantreference)
* [catalogv1alpha1.TenantReferenceList](#catalogv1alpha1.tenantreferencelist)
* [catalogv1alpha1.TenantReferenceSpec](#catalogv1alpha1.tenantreferencespec)
* [operatorv1alpha1.Catapult](#operatorv1alpha1.catapult)
* [operatorv1alpha1.CatapultCondition](#operatorv1alpha1.catapultcondition)
* [operatorv1alpha1.CatapultList](#operatorv1alpha1.catapultlist)
* [operatorv1alpha1.CatapultSpec](#operatorv1alpha1.catapultspec)
* [operatorv1alpha1.CatapultStatus](#operatorv1alpha1.catapultstatus)
* [operatorv1alpha1.Elevator](#operatorv1alpha1.elevator)
* [operatorv1alpha1.ElevatorCondition](#operatorv1alpha1.elevatorcondition)
* [operatorv1alpha1.ElevatorList](#operatorv1alpha1.elevatorlist)
* [operatorv1alpha1.ElevatorSpec](#operatorv1alpha1.elevatorspec)
* [operatorv1alpha1.ElevatorStatus](#operatorv1alpha1.elevatorstatus)
* [operatorv1alpha1.KubeCarrier](#operatorv1alpha1.kubecarrier)
* [operatorv1alpha1.KubeCarrierCondition](#operatorv1alpha1.kubecarriercondition)
* [operatorv1alpha1.KubeCarrierList](#operatorv1alpha1.kubecarrierlist)
* [operatorv1alpha1.KubeCarrierSpec](#operatorv1alpha1.kubecarrierspec)
* [operatorv1alpha1.KubeCarrierStatus](#operatorv1alpha1.kubecarrierstatus)
* [operatorv1alpha1.ServiceClusterRegistration](#operatorv1alpha1.serviceclusterregistration)
* [operatorv1alpha1.ServiceClusterRegistrationCondition](#operatorv1alpha1.serviceclusterregistrationcondition)
* [operatorv1alpha1.ServiceClusterRegistrationList](#operatorv1alpha1.serviceclusterregistrationlist)
* [operatorv1alpha1.ServiceClusterRegistrationSpec](#operatorv1alpha1.serviceclusterregistrationspec)
* [operatorv1alpha1.ServiceClusterRegistrationStatus](#operatorv1alpha1.serviceclusterregistrationstatus)
* [operatorv1alpha1.CRDReference](#operatorv1alpha1.crdreference)
* [operatorv1alpha1.ObjectReference](#operatorv1alpha1.objectreference)
* [corev1alpha1.CustomResourceDiscovery](#corev1alpha1.customresourcediscovery)
* [corev1alpha1.CustomResourceDiscoveryCondition](#corev1alpha1.customresourcediscoverycondition)
* [corev1alpha1.CustomResourceDiscoveryList](#corev1alpha1.customresourcediscoverylist)
* [corev1alpha1.CustomResourceDiscoverySpec](#corev1alpha1.customresourcediscoveryspec)
* [corev1alpha1.CustomResourceDiscoveryStatus](#corev1alpha1.customresourcediscoverystatus)
* [corev1alpha1.CustomResourceDiscoverySet](#corev1alpha1.customresourcediscoveryset)
* [corev1alpha1.CustomResourceDiscoverySetCondition](#corev1alpha1.customresourcediscoverysetcondition)
* [corev1alpha1.CustomResourceDiscoverySetList](#corev1alpha1.customresourcediscoverysetlist)
* [corev1alpha1.CustomResourceDiscoverySetSpec](#corev1alpha1.customresourcediscoverysetspec)
* [corev1alpha1.CustomResourceDiscoverySetStatus](#corev1alpha1.customresourcediscoverysetstatus)
* [corev1alpha1.ServiceCluster](#corev1alpha1.servicecluster)
* [corev1alpha1.ServiceClusterCondition](#corev1alpha1.serviceclustercondition)
* [corev1alpha1.ServiceClusterList](#corev1alpha1.serviceclusterlist)
* [corev1alpha1.ServiceClusterMetadata](#corev1alpha1.serviceclustermetadata)
* [corev1alpha1.ServiceClusterSpec](#corev1alpha1.serviceclusterspec)
* [corev1alpha1.ServiceClusterStatus](#corev1alpha1.serviceclusterstatus)
* [corev1alpha1.ServiceClusterAssignment](#corev1alpha1.serviceclusterassignment)
* [corev1alpha1.ServiceClusterAssignmentCondition](#corev1alpha1.serviceclusterassignmentcondition)
* [corev1alpha1.ServiceClusterAssignmentList](#corev1alpha1.serviceclusterassignmentlist)
* [corev1alpha1.ServiceClusterAssignmentSpec](#corev1alpha1.serviceclusterassignmentspec)
* [corev1alpha1.ServiceClusterAssignmentStatus](#corev1alpha1.serviceclusterassignmentstatus)
* [corev1alpha1.ObjectReference](#corev1alpha1.objectreference)

## catalogv1alpha1.ObjectReference

ObjectReference describes the link to another object in the same namespace.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.Catalog

Catalog publishes a selection of CatalogEntries to a selection of Tenants.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.CatalogSpec](#catalogv1alpha1.catalogspec) | false |
| status |  | [catalogv1alpha1.CatalogStatus](#catalogv1alpha1.catalogstatus) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogCondition

CatalogCondition contains details for the current condition of this Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catalog condition, currently ('Ready'). | catalogv1alpha1.CatalogConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalogv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogList

CatalogList contains a list of Catalog

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.Catalog](#catalogv1alpha1.catalog) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogSpec

CatalogSpec defines the desired contents of a Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| catalogEntrySelector | CatalogEntrySelector selects CatalogEntry objects that should be part of this catalog. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |
| tenantReferenceSelector | TenantReferenceSelector selects TenantReference objects that the catalog should be published to. | *[metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogStatus

CatalogStatus defines the observed state of Catalog.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| tenants | Tenants is the list of the Tenants(TenantReference) that selected by this Catalog. | [][catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | false |
| entries | Entries is the list of the CatalogEntries that selected by this Catalog. | [][catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catalog by the controller. | catalogv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catalog's current state. | [][catalogv1alpha1.CatalogCondition](#catalogv1alpha1.catalogcondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalogv1alpha1.CatalogPhaseType | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntry

CatalogEntry reference to the CRD that the provider wants to provide as service to the tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.CatalogEntrySpec](#catalogv1alpha1.catalogentryspec) | false |
| status |  | [catalogv1alpha1.CatalogEntryStatus](#catalogv1alpha1.catalogentrystatus) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntryCondition

CatalogEntryCondition contains details for the current condition of this CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the CatalogEntry condition, currently ('Ready'). | catalogv1alpha1.CatalogEntryConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalogv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntryList

CatalogEntryList contains a list of CatalogEntry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.CatalogEntry](#catalogv1alpha1.catalogentry) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntryMetadata

CatalogEntryMetadata contains the metadata (display name, description, etc) of the CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this CatalogEntry. | string | false |
| description | Description shows the human-readable description of this CatalogEntry. | string | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntrySpec

CatalogEntrySpec defines the desired state of CatalogEntry

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the CatalogEntry. | [catalogv1alpha1.CatalogEntryMetadata](#catalogv1alpha1.catalogentrymetadata) | false |
| referencedCRD | ReferencedCRD is the underlying ReferencedCRD objects that this CatalogEntry refers to. | [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CatalogEntryStatus

CatalogEntryStatus defines the observed state of CatalogEntry.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD holds the information about the underlying ReferencedCRD that are offered by this CatalogEntry. | [catalogv1alpha1.CRDInformation](#catalogv1alpha1.crdinformation) | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this CatalogEntry by the controller. | catalogv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a CatalogEntry's current state. | [][catalogv1alpha1.CatalogEntryCondition](#catalogv1alpha1.catalogentrycondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalogv1alpha1.CatalogEntryPhaseType | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CRDInformation

CRDInformation contains type information about the CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |
| apiGroup |  | string | true |
| kind |  | string | true |
| versions |  | [][catalogv1alpha1.CRDVersion](#catalogv1alpha1.crdversion) | true |
| serviceCluster | ServiceCluster references a ServiceClusterReference of this CRD. | [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.CRDVersion

CRDVersion holds CRD version specific details.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of this version, for example: v1, v1alpha1, v1beta1 | string | true |
| schema | Schema of this CRD version. | *apiextensionsv1.CustomResourceValidation | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResource

DerivedCustomResource derives a new CRD from a existing one.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.DerivedCustomResourceSpec](#catalogv1alpha1.derivedcustomresourcespec) | false |
| status |  | [catalogv1alpha1.DerivedCustomResourceStatus](#catalogv1alpha1.derivedcustomresourcestatus) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResourceCondition

DerivedCustomResourceCondition contains details for the current condition of this DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the DerivedCustomResource condition, currently ('Ready'). | catalogv1alpha1.DerivedCustomResourceConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalogv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResourceList

DerivedCustomResourceList contains a list of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.DerivedCustomResource](#catalogv1alpha1.derivedcustomresource) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResourceReference

DerivedCustomResourceReference references the derived CRD controlled by this DerivedCustomResource instance.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Name of the derived CRD. | string | true |
| group | API Group of the derived CRD. | string | true |
| kind |  | string | true |
| plural |  | string | true |
| singular |  | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResourceSpec

DerivedCustomResourceSpec defines the desired state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| baseCRD | CRD that should be used as a base to derive a new CRD from. | [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | true |
| kindOverride | overrides the kind of the derived CRD. | string | false |
| expose | controls which fields will be present in the derived CRD. | [][catalogv1alpha1.VersionExposeConfig](#catalogv1alpha1.versionexposeconfig) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.DerivedCustomResourceStatus

DerivedCustomResourceStatus defines the observed state of DerivedCustomResource.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this DerivedCustomResource by the controller. | catalogv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a DerivedCustomResource's current state. | [][catalogv1alpha1.DerivedCustomResourceCondition](#catalogv1alpha1.derivedcustomresourcecondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalogv1alpha1.DerivedCustomResourcePhaseType | false |
| derivedCR | DerivedCR holds information about the derived CRD. | *[catalogv1alpha1.DerivedCustomResourceReference](#catalogv1alpha1.derivedcustomresourcereference) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.FieldPath

FieldPath is specifying how to address a certain field.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| jsonPath | JSONPath e.g. .spec.somefield.somesubfield | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.VersionExposeConfig

VersionExposeConfig specifies which fields to expose in the derived CRD.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| versions | specifies the versions of the referenced CRD, that this expose config applies to. The same version may not be specified in multiple VersionExposeConfigs. | []string | true |
| fields | specifies the fields that should be present in the derived CRD. | [][catalogv1alpha1.FieldPath](#catalogv1alpha1.fieldpath) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.Offering

Offering is used for Tenants to discover services that have been made available to them.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| offering |  | [catalogv1alpha1.OfferingData](#catalogv1alpha1.offeringdata) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.OfferingData

OfferingData defines the data (metadata, provider, crds, etc.) of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [catalogv1alpha1.OfferingMetadata](#catalogv1alpha1.offeringmetadata) | false |
| provider | Provider references a ProviderReference of this Offering. | [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | true |
| crd | CRD holds the information about the underlying CRD that is offered by this offering. | [catalogv1alpha1.CRDInformation](#catalogv1alpha1.crdinformation) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.OfferingList

OfferingList contains a list of Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.Offering](#catalogv1alpha1.offering) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.OfferingMetadata

OfferingMetadata contains the metadata (display name, description, etc) of the Offering.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this Offering. | string | false |
| description | Description shows the human-readable description of this Offering. | string | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.Provider

Provider is the service provider representation in the KubeCarrier control-plane.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.ProviderSpec](#catalogv1alpha1.providerspec) | false |
| status |  | [catalogv1alpha1.ProviderStatus](#catalogv1alpha1.providerstatus) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderCondition

ProviderCondition contains details for the current condition of this Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Provider condition, currently ('Ready'). | catalogv1alpha1.ProviderConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalogv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderList

ProviderList contains a list of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.Provider](#catalogv1alpha1.provider) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderMetadata

ProviderMetadata contains the metadata (display name, description, etc) of the Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this Provider. | string | false |
| description | Description shows the human-readable description of this Provider. | string | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderSpec

ProviderSpec defines the desired state of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [catalogv1alpha1.ProviderMetadata](#catalogv1alpha1.providermetadata) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderStatus

ProviderStatus defines the observed state of Provider.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespaceName | NamespaceName is the name of the namespace that the Provider manages. | string | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Provider by the controller. | catalogv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Provider's current state. | [][catalogv1alpha1.ProviderCondition](#catalogv1alpha1.providercondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalogv1alpha1.ProviderPhaseType | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderReference

ProviderReference exposes information of the Provider(displayName, description). This object lives in the tenant namespace for each provider the tenant is allowed utilizing (e.g. there's catalog selecting this tenant as its user)

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.ProviderReferenceSpec](#catalogv1alpha1.providerreferencespec) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderReferenceList

ProviderReferenceList contains a list of ProviderReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.ProviderReference](#catalogv1alpha1.providerreference) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ProviderReferenceSpec

ProviderReferenceSpec defines the desired state of ProviderReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the Provider. | [catalogv1alpha1.ProviderMetadata](#catalogv1alpha1.providermetadata) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ServiceClusterReference

ServiceClusterReference exposes information about a Providers Clusters.\n\nThis object lives in the tenant namespace for each provider the tenant is allowed utilizing (e.g. there's catalog selecting this tenant as its user)

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.ServiceClusterReferenceSpec](#catalogv1alpha1.serviceclusterreferencespec) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ServiceClusterReferenceList

ServiceClusterReferenceList contains a list of ServiceClusterReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.ServiceClusterReference](#catalogv1alpha1.serviceclusterreference) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.ServiceClusterReferenceSpec

ServiceClusterReferenceSpec defines the desired state of ServiceClusterReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Metadata contains the metadata (display name, description, etc) of the ServiceCluster. | [corev1alpha1.ServiceClusterMetadata](#corev1alpha1.serviceclustermetadata) | false |
| provider | Provider references the Provider that this ServiceCluster belongs to. | [catalogv1alpha1.ObjectReference](#catalogv1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.Tenant

Tenant sets up permissions and references to allow a end-user group to interact with providers' services.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.TenantSpec](#catalogv1alpha1.tenantspec) | false |
| status |  | [catalogv1alpha1.TenantStatus](#catalogv1alpha1.tenantstatus) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantCondition

TenantCondition contains details for the current condition of this Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Tenant condition, currently ('Ready'). | catalogv1alpha1.TenantConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | catalogv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantList

TenantList contains a list of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.Tenant](#catalogv1alpha1.tenant) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantSpec

TenantSpec defines the desired state of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantStatus

TenantStatus defines the observed state of Tenant.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| namespaceName | NamespaceName is the name of the namespace that the Tenant manages. | string | false |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Tenant by the controller. | catalogv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Tenant's current state. | [][catalogv1alpha1.TenantCondition](#catalogv1alpha1.tenantcondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | catalogv1alpha1.TenantPhaseType | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantReference

TenantReference is a read-only object exposing the Tenant information. TenantReference lives in the provider's namespace. The provider is allowed modifying TenantReference's labels, marking them at will. This allows the tenant granular tenant selection for the offered services catalogs.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [catalogv1alpha1.TenantReferenceSpec](#catalogv1alpha1.tenantreferencespec) | false |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantReferenceList

TenantReferenceList contains a list of TenantReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][catalogv1alpha1.TenantReference](#catalogv1alpha1.tenantreference) | true |

[Back to TOC](#table-of-contents)

## catalogv1alpha1.TenantReferenceSpec

TenantReferenceSpec defines the desired state of TenantReference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.Catapult

Catapult manages the deployment of the Catapult controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [operatorv1alpha1.CatapultSpec](#operatorv1alpha1.catapultspec) | false |
| status |  | [operatorv1alpha1.CatapultStatus](#operatorv1alpha1.catapultstatus) | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.CatapultCondition

CatapultCondition contains details for the current condition of this Catapult.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Catapult condition, currently ('Ready'). | operatorv1alpha1.CatapultConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operatorv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.CatapultList

CatapultList contains a list of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][operatorv1alpha1.Catapult](#operatorv1alpha1.catapult) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.CatapultSpec

CatapultSpec defines the desired state of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| masterClusterCRD | References the CRD in the master cluster. | [operatorv1alpha1.CRDReference](#operatorv1alpha1.crdreference) | true |
| serviceClusterCRD | References the CRD in the service cluster. | [operatorv1alpha1.CRDReference](#operatorv1alpha1.crdreference) | true |
| serviceCluster | References the ServiceCluster object that this object belongs to. | [operatorv1alpha1.ObjectReference](#operatorv1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.CatapultStatus

CatapultStatus defines the observed state of Catapult

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Catapult by the controller. | operatorv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Catapult's current state. | [][operatorv1alpha1.CatapultCondition](#operatorv1alpha1.catapultcondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operatorv1alpha1.CatapultPhaseType | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.Elevator

Elevator manages the deployment of the Elevator controller manager. For each `DerivedCustomResource` a Elevator instance is launched to propagate the derived CRD instance into the provider namespace. This component works hand-in-hand with the Catapult instance for the respective type.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [operatorv1alpha1.ElevatorSpec](#operatorv1alpha1.elevatorspec) | false |
| status |  | [operatorv1alpha1.ElevatorStatus](#operatorv1alpha1.elevatorstatus) | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ElevatorCondition

ElevatorCondition contains details for the current condition of this Elevator.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the Elevator condition, currently ('Ready'). | operatorv1alpha1.ElevatorConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operatorv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ElevatorList

ElevatorList contains a list of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][operatorv1alpha1.Elevator](#operatorv1alpha1.elevator) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ElevatorSpec

ElevatorSpec defines the desired state of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| providerCRD | References the provider or internal CRD, that should be created in the provider namespace. | [operatorv1alpha1.CRDReference](#operatorv1alpha1.crdreference) | true |
| tenantCRD | References the public CRD that will be synced into the provider namespace. | [operatorv1alpha1.CRDReference](#operatorv1alpha1.crdreference) | true |
| derivedCR | References the DerivedCustomResource controlling the Tenant-side CRD. | [operatorv1alpha1.ObjectReference](#operatorv1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ElevatorStatus

ElevatorStatus defines the observed state of Elevator

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this Elevator by the controller. | operatorv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a Elevator's current state. | [][operatorv1alpha1.ElevatorCondition](#operatorv1alpha1.elevatorcondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operatorv1alpha1.ElevatorPhaseType | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.KubeCarrier

KubeCarrier manages the deployment of the KubeCarrier controller manager.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [operatorv1alpha1.KubeCarrierSpec](#operatorv1alpha1.kubecarrierspec) | false |
| status |  | [operatorv1alpha1.KubeCarrierStatus](#operatorv1alpha1.kubecarrierstatus) | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.KubeCarrierCondition

KubeCarrierCondition contains details for the current condition of this KubeCarrier.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| type | Type is the type of the KubeCarrier condition, currently ('Ready'). | operatorv1alpha1.KubeCarrierConditionType | true |
| status | Status is the status of the condition, one of ('True', 'False', 'Unknown'). | operatorv1alpha1.ConditionStatus | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transits from one status to another. | metav1.Time | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| message | Message is the human readable message indicating details about last transition. | string | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.KubeCarrierList

KubeCarrierList contains a list of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][operatorv1alpha1.KubeCarrier](#operatorv1alpha1.kubecarrier) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.KubeCarrierSpec

KubeCarrierSpec defines the desired state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.KubeCarrierStatus

KubeCarrierStatus defines the observed state of KubeCarrier

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| observedGeneration | ObservedGeneration is the most recent generation observed for this KubeCarrier by the controller. | operatorv1alpha1.int64 | false |
| conditions | Conditions represents the latest available observations of a KubeCarrier's current state. | [][operatorv1alpha1.KubeCarrierCondition](#operatorv1alpha1.kubecarriercondition) | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operatorv1alpha1.KubeCarrierPhaseType | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ServiceClusterRegistration

ServiceClusterRegistration represents single kubernetes cluster belonging to the provider\n\nServiceClusterRegistration lives in the provider namespace. For each ferry the kubecarrier operator spins up the ferry controller deployment, necessary roles, service accounts, and role bindings\n\nThe reason for ferry controller deployment are multiples: * security --> kubecarrier operator has greater privileges then ferry controller * resource isolation --> each ferry controller pod operates only on a single service cluster,\n\t\tthus resource allocation and monitoring is separate per ferrys. This allows finer grade\n\t\tresource tuning and monitoring\n* flexibility --> If needed different ferrys could have different deployments depending on\n\t\ttheir specific need (e.g. kubecarrier image version for gradual rolling upgrade, different resource allocation, etc),

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [operatorv1alpha1.ServiceClusterRegistrationSpec](#operatorv1alpha1.serviceclusterregistrationspec) | false |
| status |  | [operatorv1alpha1.ServiceClusterRegistrationStatus](#operatorv1alpha1.serviceclusterregistrationstatus) | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ServiceClusterRegistrationCondition

ServiceClusterRegistrationCondition contains details for the current condition of this ServiceClusterRegistration.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | operatorv1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | operatorv1alpha1.ServiceClusterRegistrationConditionType | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ServiceClusterRegistrationList

ServiceClusterRegistrationList contains a list of ServiceClusterRegistration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][operatorv1alpha1.ServiceClusterRegistration](#operatorv1alpha1.serviceclusterregistration) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ServiceClusterRegistrationSpec

ServiceClusterRegistrationSpec defines the desired state of ServiceClusterRegistration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kubeconfigSecret | KubeconfigSecret specifies the Kubeconfig to use when connecting to the ServiceCluster. | [operatorv1alpha1.ObjectReference](#operatorv1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ServiceClusterRegistrationStatus

ServiceClusterRegistrationStatus defines the observed state of ServiceClusterRegistration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object. Consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to strings when printing the property. This is only for display purpose, for everything else use conditions. | operatorv1alpha1.ServiceClusterRegistrationPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceClusterRegistration is in. | [][operatorv1alpha1.ServiceClusterRegistrationCondition](#operatorv1alpha1.serviceclusterregistrationcondition) | false |
| observedGeneration | The most recent generation observed by the controller. | operatorv1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.CRDReference

CRDReference references a CustomResourceDefitition.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| kind |  | string | true |
| version |  | string | true |
| group |  | string | true |
| plural |  | string | true |

[Back to TOC](#table-of-contents)

## operatorv1alpha1.ObjectReference

ObjectReference describes the link to another object in the same namespace

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscovery

CustomResourceDiscovery is used inside KubeCarrier to fetch a CustomResourceDefinition from another cluster and to offload cross cluster access to another component.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [corev1alpha1.CustomResourceDiscoverySpec](#corev1alpha1.customresourcediscoveryspec) | false |
| status |  | [corev1alpha1.CustomResourceDiscoveryStatus](#corev1alpha1.customresourcediscoverystatus) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoveryCondition

CustomResourceDiscoveryCondition contains details for the current condition of this CustomResourceDiscovery.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | corev1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | corev1alpha1.CustomResourceDiscoveryConditionType | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoveryList

CustomResourceDiscoveryList contains a list of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][corev1alpha1.CustomResourceDiscovery](#corev1alpha1.customresourcediscovery) | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySpec

CustomResourceDiscoverySpec defines the desired state of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | true |
| serviceCluster | ServiceCluster references a ServiceCluster to search the CustomResourceDefinition on. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | true |
| kindOverride | KindOverride overrides resulting internal CRDs kind | string | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoveryStatus

CustomResourceDiscoveryStatus defines the observed state of crdreference

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD defines the original CustomResourceDefinition specification from the service cluster | *apiextensionsv1.CustomResourceDefinition | false |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | corev1alpha1.CustomResourceDiscoveryPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | [][corev1alpha1.CustomResourceDiscoveryCondition](#corev1alpha1.customresourcediscoverycondition) | false |
| observedGeneration | The most recent generation observed by the controller. | corev1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySet

CustomResourceDiscoverySet manages multiple CustomResourceDiscovery objects for a set of service clusters.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [corev1alpha1.CustomResourceDiscoverySetSpec](#corev1alpha1.customresourcediscoverysetspec) | false |
| status |  | [corev1alpha1.CustomResourceDiscoverySetStatus](#corev1alpha1.customresourcediscoverysetstatus) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySetCondition

CustomResourceDiscoverySetCondition contains details for the current condition of this CustomResourceDiscoverySet.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | corev1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | corev1alpha1.CustomResourceDiscoverySetConditionType | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySetList



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][corev1alpha1.CustomResourceDiscoverySet](#corev1alpha1.customresourcediscoveryset) | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySetSpec



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| crd | CRD references a CustomResourceDefinition within the ServiceCluster. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | true |
| serviceClusterSelector | ServiceClusterSelector references a set of ServiceClusters to search the CustomResourceDefinition on. | [metav1.LabelSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#labelselector-v1-meta) | true |
| kindOverride | KindOverride overrides resulting internal CRDs kind | string | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.CustomResourceDiscoverySetStatus



| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | corev1alpha1.CustomResourceDiscoverySetPhaseType | false |
| conditions | Conditions is a list of all conditions this CustomResourceDiscovery is in. | [][corev1alpha1.CustomResourceDiscoverySetCondition](#corev1alpha1.customresourcediscoverysetcondition) | false |
| observedGeneration | The most recent generation observed by the controller. | corev1alpha1.int64 | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceCluster

ServiceCluster is a providers Kubernetes Cluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [corev1alpha1.ServiceClusterSpec](#corev1alpha1.serviceclusterspec) | false |
| status |  | [corev1alpha1.ServiceClusterStatus](#corev1alpha1.serviceclusterstatus) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterCondition

ServiceClusterCondition contains details for the current condition of this ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastHeartbeatTime | LastHeartbeatTime is the timestamp corresponding to the last update of this condition. | metav1.Time | true |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | corev1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | corev1alpha1.ServiceClusterConditionType | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterList

ServiceClusterList contains a list of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][corev1alpha1.ServiceCluster](#corev1alpha1.servicecluster) | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterMetadata

ServiceClusterMetadata contains the metadata (display name, description, etc) of the ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| displayName | DisplayName shows the human-readable name of this ServiceCluster. | string | false |
| description | Description shows the human-readable description of this ServiceCluster. | string | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterSpec

ServiceClusterSpec defines the desired state of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [corev1alpha1.ServiceClusterMetadata](#corev1alpha1.serviceclustermetadata) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterStatus

ServiceClusterStatus defines the observed state of ServiceCluster

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | corev1alpha1.ServiceClusterPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceCluster is in. | [][corev1alpha1.ServiceClusterCondition](#corev1alpha1.serviceclustercondition) | false |
| observedGeneration | The most recent generation observed by the controller. | corev1alpha1.int64 | false |
| kubernetesVersion | KubernetesVersion of the service cluster API Server | *version.Info | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterAssignment

ServiceClusterAssignment represents the assignment of a Tenant to a ServiceCluster.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#objectmeta-v1-meta) | false |
| spec |  | [corev1alpha1.ServiceClusterAssignmentSpec](#corev1alpha1.serviceclusterassignmentspec) | false |
| status |  | [corev1alpha1.ServiceClusterAssignmentStatus](#corev1alpha1.serviceclusterassignmentstatus) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterAssignmentCondition

ServiceClusterAssignmentCondition contains details for the current condition of this ServiceClusterAssignment.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| lastTransitionTime | LastTransitionTime is the last time the condition transit from one status to another. | metav1.Time | true |
| message | Message is the human readable message indicating details about last transition. | string | true |
| reason | Reason is the (brief) reason for the condition's last transition. | string | true |
| status | Status of the condition, one of ('True', 'False', 'Unknown'). | corev1alpha1.ConditionStatus | true |
| type | Type of the condition, currently ('Ready'). | corev1alpha1.ServiceClusterAssignmentConditionType | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterAssignmentList

ServiceClusterAssignmentList contains a list of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.14/#listmeta-v1-meta) | false |
| items |  | [][corev1alpha1.ServiceClusterAssignment](#corev1alpha1.serviceclusterassignment) | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterAssignmentSpec

ServiceClusterAssignmentSpec defines the desired state of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| serviceCluster | References the ServiceCluster. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | true |
| masterNamespace | References the source namespace in the master cluster. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | true |

[Back to TOC](#table-of-contents)

## corev1alpha1.ServiceClusterAssignmentStatus

ServiceClusterAssignmentStatus defines the observed state of ServiceClusterAssignment

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| phase | DEPRECATED. Phase represents the current lifecycle state of this object consider this field DEPRECATED, it will be removed as soon as there is a mechanism to map conditions to a string when printing the property is only present for display purposes, for everything else use conditions | corev1alpha1.ServiceClusterAssignmentPhaseType | false |
| conditions | Conditions is a list of all conditions this ServiceClusterAssignment is in. | [][corev1alpha1.ServiceClusterAssignmentCondition](#corev1alpha1.serviceclusterassignmentcondition) | false |
| observedGeneration | The most recent generation observed by the controller. | corev1alpha1.int64 | false |
| serviceClusterNamespace | ServiceClusterNamespace references the Namespace on the ServiceCluster that was assigned. | [corev1alpha1.ObjectReference](#corev1alpha1.objectreference) | false |

[Back to TOC](#table-of-contents)

## corev1alpha1.ObjectReference

ObjectReference describes the link to another object in the same namespace

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name |  | string | true |

[Back to TOC](#table-of-contents)
