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

package util

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type object interface {
	metav1.Object
	runtime.Object
}

// EnsureUniqueNamespace generates unique namespace for obj if one already doesn't exists
//
// It's required that OwnerReverseFieldIndex exists for corev1.Namespace
func EnsureUniqueNamespace(ctx context.Context, c client.Client, scheme *runtime.Scheme, owner object) (*corev1.Namespace, error) {
	ownedBy, err := OwnedBy(owner, scheme)
	if err != nil {
		return nil, fmt.Errorf("building owned by selector: %w", err)
	}
	namespaceList := &corev1.NamespaceList{}
	if err = c.List(ctx, namespaceList, ownedBy); err != nil {
		return nil, fmt.Errorf("listing Namespaces: %w", err)
	}

	switch len(namespaceList.Items) {
	case 0:
		// Create Namespace
		gvk, err := apiutil.GVKForObject(owner, scheme)
		if err != nil {
			return nil, fmt.Errorf("kind from scheme: %w", err)
		}
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: strings.ToLower(gvk.Kind) + "-",
			},
		}
		if _, err := InsertOwnerReference(owner, namespace, scheme); err != nil {
			return nil, fmt.Errorf("inserting Owner Reference: %w", err)
		}
		if err = c.Create(ctx, namespace); err != nil {
			return nil, fmt.Errorf("creating Namespace: %w", err)
		}
		return namespace, nil
	case 1:
		return namespaceList.Items[0].DeepCopy(), nil
	default:
		nss := make([]string, len(namespaceList.Items))
		for i, ns := range namespaceList.Items {
			nss[i] = ns.Name
		}
		return nil, fmt.Errorf("MultipleNamespaces owned namespaces found: %s", strings.Join(nss, ","))
	}
}
