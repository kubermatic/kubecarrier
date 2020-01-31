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
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ProviderNamespaceFieldIndex = "provider.kubecarrier.io/namespace"
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

func GetProviderByProviderNamespace(ctx context.Context, c client.Client, kubecarrierNamespace, providerNamespace string) (*Provider, error) {
	providerList := &ProviderList{}
	if err := c.List(ctx, providerList,
		client.InNamespace(kubecarrierNamespace),
		client.MatchingFields{
			ProviderNamespaceFieldIndex: providerNamespace,
		},
	); err != nil {
		return nil, err
	}
	switch len(providerList.Items) {
	case 0:
		// not found
		return nil, fmt.Errorf("providers.catalog.kubecarrier.io with index %q not found", ProviderNamespaceFieldIndex)
	case 1:
		// found!
		return &providerList.Items[0], nil
	default:
		// found too many
		return nil, fmt.Errorf("multiple providers.catalog.kubecarrier.io with index %q found", ProviderNamespaceFieldIndex)
	}
}
