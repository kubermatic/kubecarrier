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

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomResourceDefinition reconciles a apps/v1, Kind=CustromResourceDefinition.
func CustomResourceDefinition(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredCRD *apiextensionsv1beta1.CustomResourceDefinition,
) (currentCRD *apiextensionsv1beta1.CustomResourceDefinition, err error) {
	name := types.NamespacedName{
		Name:      desiredCRD.Name,
		Namespace: desiredCRD.Namespace,
	}

	// Lookup current version of the object
	currentCRD = &apiextensionsv1beta1.CustomResourceDefinition{}
	err = c.Get(ctx, name, currentCRD)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting CRD: %w", err)
	}

	if errors.IsNotFound(err) {
		// CRD needs to be created
		log.V(1).Info("creating", "CRD", name.String())
		if err = c.Create(ctx, desiredCRD); err != nil {
			return nil, fmt.Errorf("creating CRD: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredCRD, nil
	}

	if !equality.Semantic.DeepEqual(desiredCRD.Spec, currentCRD.Spec) {
		// desired and current CRD .Spec are not equal -> trigger an update
		log.V(1).Info("updating", "CRD", name.String())
		currentCRD.Spec = desiredCRD.Spec
		if err = c.Update(ctx, currentCRD); err != nil {
			return nil, fmt.Errorf("updating CRD: %w", err)
		}
	}

	return currentCRD, nil
}
