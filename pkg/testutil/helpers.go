/*
Copyright 2019 The KubeCarrier Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	corev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/core/v1alpha1"
	fakev1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/constants"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

func ConditionStatusEqual(obj runtime.Object, ConditionType, ConditionStatus interface{}) error {
	jp := jsonpath.New("condition")
	if err := jp.Parse(fmt.Sprintf(`{.status.conditions[?(@.type=="%s")].status}`, ConditionType)); err != nil {
		return err
	}
	res, err := jp.FindResults(obj)
	if err != nil {
		return fmt.Errorf("cannot find results: %w", err)
	}
	if len(res) != 1 {
		return fmt.Errorf("found %d matching conditions, expected 1", len(res))
	}
	rr := res[0]
	if len(rr) != 1 {
		return fmt.Errorf("found %d matching conditions, expected 1", len(rr))
	}
	status := rr[0].String()
	if status != fmt.Sprint(ConditionStatus) {
		return fmt.Errorf("expected condition status %s, got %s", ConditionStatus, status)
	}
	return nil
}

func LogObject(t *testing.T, obj interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(obj, "", "\t")
	require.NoError(t, err)
	t.Log("\n", string(b))
}

var (
	WithTimeout = util.WithClientWatcherTimeout
)

func WaitUntilNotFound(ctx context.Context, c *RecordingClient, obj runtime.Object, options ...util.ClientWatcherOption) error {
	c.t.Helper()
	return c.WaitUntilNotFound(ctx, obj, options...)
}

func WaitUntilFound(ctx context.Context, c *RecordingClient, obj runtime.Object, options ...util.ClientWatcherOption) error {
	c.t.Helper()
	return c.WaitUntil(ctx, obj, func() (done bool, err error) {
		return true, nil
	}, options...)
}

func WaitUntilCondition(ctx context.Context, c *RecordingClient, obj runtime.Object, ConditionType, conditionStatus interface{}, options ...util.ClientWatcherOption) error {
	c.t.Helper()
	err := c.WaitUntil(ctx, obj, func() (done bool, err error) {
		return ConditionStatusEqual(obj, ConditionType, conditionStatus) == nil, nil
	}, options...)

	if err != nil {
		b, marshallErr := json.MarshalIndent(obj, "", "\t")
		if marshallErr != nil {
			return fmt.Errorf("cannot marshall indent obj!!! %v %w", marshallErr, err)
		}
		return fmt.Errorf("%w\n%s", err, string(b))
	}
	return nil
}

func WaitUntilReady(ctx context.Context, c *RecordingClient, obj runtime.Object, options ...util.ClientWatcherOption) error {
	c.t.Helper()
	return WaitUntilCondition(ctx, c, obj, "Ready", "True", options...)
}

func DeleteAndWaitUntilNotFound(ctx context.Context, c *RecordingClient, obj runtime.Object, options ...util.ClientWatcherOption) error {
	c.t.Helper()
	if err := c.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
		return err
	}
	return WaitUntilNotFound(ctx, c, obj, options...)
}

func KubeCarrierOperatorCheck(
	ctx context.Context,
	t *testing.T,
	managementClient *RecordingClient,
	managementScheme *runtime.Scheme,
) {
	namePrefix := constants.KubeCarrierDefaultName + "-operator"
	namespaceName := constants.KubeCarrierDefaultNamespace

	componentCheck(ctx, managementClient, managementScheme, t, namePrefix, []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-manager", namePrefix),
				Namespace: namespaceName,
			},
		},
		// Webhook Service
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-webhook-service", namePrefix),
				Namespace: namespaceName,
			},
		},
		// ClusterRoleBinding
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-rolebinding", namePrefix),
			},
		},
		// ClusterRole
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-role", namePrefix),
			},
		},
		// RoleBinding
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-leader-election-rolebinding", namePrefix),
				Namespace: namespaceName,
			},
		},
		// Role
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-leader-election-role", namePrefix),
				Namespace: namespaceName,
			},
		},
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "accounts.catalog.kubecarrier.io",
			},
		},
	})
}

func E2EOperatorCheck(
	ctx context.Context,
	t *testing.T,
	serviceClient *RecordingClient,
	serviceScheme *runtime.Scheme,
) {
	namePrefix := "e2e-operator"
	namespaceName := "kubecarrier-e2e-operator"

	componentCheck(ctx, serviceClient, serviceScheme, t, namePrefix, []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-manager", namePrefix),
				Namespace: namespaceName,
			},
		},
		// ClusterRoleBinding
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-rolebinding", namePrefix),
			},
		},
		// ClusterRole
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-role", namePrefix),
			},
		},
		&apiextensionsv1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "dbs.fake.kubecarrier.io",
			},
		},
	})
}

func KubeCarrierCheck(
	ctx context.Context,
	t *testing.T,
	managementClient *RecordingClient,
	managementScheme *runtime.Scheme,
) {
	namePrefix := constants.KubeCarrierDefaultName + "-manager"
	namespaceName := constants.KubeCarrierDefaultNamespace

	componentCheck(ctx, managementClient, managementScheme, t, namePrefix, []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespaceName,
			},
		},
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-controller-manager", namePrefix),
				Namespace: namespaceName,
			},
		},
		// Webhook Service
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-webhook-service", namePrefix),
				Namespace: namespaceName,
			},
		},
		// ClusterRoleBinding
		&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-rolebinding", namePrefix),
			},
		},
		// ClusterRole
		&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("%s-manager-role", namePrefix),
			},
		},
		// RoleBinding
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-leader-election-rolebinding", namePrefix),
				Namespace: namespaceName,
			},
		},
		// Role
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-leader-election-role", namePrefix),
				Namespace: namespaceName,
			},
		},
	})
}

func componentCheck(
	ctx context.Context,
	cli *RecordingClient,
	scheme *runtime.Scheme,
	t *testing.T,
	componentName string,
	ownedObjects []runtime.Object,
) {
	for _, ownedObject := range ownedObjects {
		gvk, err := apiutil.GVKForObject(ownedObject, scheme)
		require.NoError(t, err, fmt.Sprintf("cannot get GVK for %T", ownedObject))
		require.NoError(t, WaitUntilFound(ctx, cli, ownedObject), fmt.Sprintf("%s: getting %s failed", componentName, gvk.Kind))
	}
	t.Logf(fmt.Sprintf("%s is ready", componentName))
}

func NewProviderAccount(name string, subjects ...rbacv1.Subject) *catalogv1alpha1.Account {
	return &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-provider",
			Labels: map[string]string{
				"test-case": name,
			},
		},
		Spec: catalogv1alpha1.AccountSpec{
			Metadata: catalogv1alpha1.AccountMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					DisplayName:      name + " provider",
					ShortDescription: name + " provider desc",
				},
			},
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.ProviderRole,
			},
			Subjects: subjects,
		},
	}
}

func NewTenantAccount(name string, subjects ...rbacv1.Subject) *catalogv1alpha1.Account {
	return &catalogv1alpha1.Account{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-tenant",
		},
		Spec: catalogv1alpha1.AccountSpec{
			Metadata: catalogv1alpha1.AccountMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					DisplayName:      name + " tenant",
					ShortDescription: name + " tenant desc",
				},
			},
			Roles: []catalogv1alpha1.AccountRole{
				catalogv1alpha1.TenantRole,
			},
			Subjects: subjects,
		},
	}
}

func NewFakeDB(name, namespace string) *fakev1alpha1.DB {
	return &fakev1alpha1.DB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: fakev1alpha1.DBSpec{
			DatabaseName: "fakeDB",
			DatabaseUser: "user",
			Config: fakev1alpha1.Config{
				Create: fakev1alpha1.OperationFlagEnabled,
			}},
	}
}

func NewServiceCluster(name, namespace, secret string) *corev1alpha1.ServiceCluster {
	return &corev1alpha1.ServiceCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1alpha1.ServiceClusterSpec{
			Metadata: corev1alpha1.ServiceClusterMetadata{
				DisplayName: name,
				Description: fmt.Sprintf("%s service cluster", name),
			},
			KubeconfigSecret: corev1alpha1.ObjectReference{
				Name: secret,
			},
		},
	}
}

func NewCatalogEntry(name, namespace, crdName string) *catalogv1alpha1.CatalogEntry {
	return &catalogv1alpha1.CatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"kubecarrier.io/test": "label",
			},
		},
		Spec: catalogv1alpha1.CatalogEntrySpec{
			Metadata: catalogv1alpha1.CatalogEntryMetadata{
				CommonMetadata: catalogv1alpha1.CommonMetadata{
					DisplayName:      fmt.Sprintf("display name for %s", name),
					ShortDescription: fmt.Sprintf("short description for %s", name),
				},
			},
			BaseCRD: catalogv1alpha1.ObjectReference{
				Name: crdName,
			},
		},
	}
}

func NewCatalog(name, namespace string, catalogEntrySelector, tenantSelector *metav1.LabelSelector) *catalogv1alpha1.Catalog {
	return &catalogv1alpha1.Catalog{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: catalogv1alpha1.CatalogSpec{
			CatalogEntrySelector: catalogEntrySelector,
			TenantSelector:       tenantSelector,
		},
	}
}

func NewFakeCouchDBCRD(group string) *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "couchdbs." + group,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural: "couchdbs",
				Kind:   "CouchDB",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Storage: true,
					Served:  true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"apiVersion": {Type: "string"},
								"kind":       {Type: "string"},
								"metadata":   {Type: "object"},
								"spec": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"prop1": {Type: "string"},
										"prop2": {Type: "string"},
									},
								},
								"status": {
									Type: "object",
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"observedGeneration": {Type: "integer"},
										"prop1":              {Type: "string"},
										"prop2":              {Type: "string"},
									},
								},
							},
							Type: "object",
						},
					},
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
				},
			},
			Scope: apiextensionsv1.NamespaceScoped,
		},
	}
}
