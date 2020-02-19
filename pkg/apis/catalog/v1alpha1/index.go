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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

func GetProviderByProviderNamespace(ctx context.Context, c client.Client, scheme *runtime.Scheme, providerNamespace string) (*Account, error) {
	ns := &corev1.Namespace{}
	if err := c.Get(ctx, types.NamespacedName{Name: providerNamespace}, ns); err != nil {
		return nil, fmt.Errorf("getting namespace: %w", err)
	}

	owners, err := util.ListOwners(ctx, c, scheme, ns)
	if err != nil {
		return nil, fmt.Errorf("listing owners: %w", err)
	}

	var provider *Account
	for _, own := range owners {
		candidate, ok := own.(*Account)
		if ok && candidate.HasRole(ProviderRole) {
			if provider != nil {
				return nil, fmt.Errorf("multiple providers owning the namespace %s", providerNamespace)
			}
			provider = candidate
		}
	}
	if provider == nil {
		return nil, fmt.Errorf("provider with namespace %s not found", providerNamespace)
	}
	return provider, nil
}
