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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProviderNamespaceFieldIndex = "provider.kubecarrier.io/namespace"
	TenantNamespaceFieldIndex   = "tenant.kubecarrier.io/namespace"
)

// RegisterProviderNamespaceFieldIndex adds a field index for Provider.Status.NamespaceName
func RegisterProviderNamespaceFieldIndex(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(
		&Provider{}, ProviderNamespaceFieldIndex,
		client.IndexerFunc(func(obj runtime.Object) []string {
			provider := obj.(*Provider)
			return []string{provider.Status.NamespaceName}
		}))
}

// RegisterTenantNamespaceFieldIndex adds a field index for Tenant.Status.NamespaceName
func RegisterTenantNamespaceFieldIndex(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(
		&Tenant{}, TenantNamespaceFieldIndex,
		client.IndexerFunc(func(obj runtime.Object) []string {
			provider := obj.(*Tenant)
			return []string{provider.Status.NamespaceName}
		}))
}
