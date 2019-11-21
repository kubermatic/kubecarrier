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

package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomResourceDefinition reconciles a apiextensions.k8s.io/v1beta1, Kind=CustomResourceDefinition.
func CustomResourceDefinition(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredCRD *apiextensionsv1beta1.CustomResourceDefinition,
) (currentCRD *apiextensionsv1beta1.CustomResourceDefinition, err error) {
	nn := types.NamespacedName{
		Name:      desiredCRD.Name,
		Namespace: desiredCRD.Namespace,
	}

	// Lookup current version of the object
	currentCRD = &apiextensionsv1beta1.CustomResourceDefinition{}
	err = c.Get(ctx, nn, currentCRD)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting CustomResourceDefinition: %w", err)
	}

	if errors.IsNotFound(err) {
		// CustomResourceDefinition needs to be created.
		log.V(1).Info("creating", "CustomResourceDefinition", nn.String())
		if err = c.Create(ctx, desiredCRD); err != nil {
			return nil, fmt.Errorf("creating CustomResourceDefinition: %w", err)
		}
		// no need to check for updates, object was created just now
		return currentCRD, nil
	}
	if !equality.Semantic.DeepEqual(desiredCRD.Spec, currentCRD.Spec) {
		// desired and current CustomResourceDefinition .Spec are not equal -> trigger an update
		log.V(1).Info("updating", "CustomResourceDefinition", nn.String())
		currentCRD.Spec = desiredCRD.Spec
		if err = c.Update(ctx, currentCRD); err != nil {
			return nil, fmt.Errorf("updating CustomResourceDefinition: %w", err)
		}
	}
	return currentCRD, nil
}
