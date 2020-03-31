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
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/operator/v1alpha1"
)

// Tower reconciles a operator.kubecarrier.io/v1alpha1, Kind=Role.
func Tower(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredTower *operatorv1alpha1.Tower,
) (currentTower *operatorv1alpha1.Tower, err error) {
	name := types.NamespacedName{
		Name:      desiredTower.Name,
		Namespace: desiredTower.Namespace,
	}

	// Lookup current version of the object
	currentTower = &operatorv1alpha1.Tower{}
	err = c.Get(ctx, name, currentTower)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting Tower: %w", err)
	}

	if errors.IsNotFound(err) {
		// Role needs to be created
		log.V(1).Info("creating", "Tower", name.String())
		if err = c.Create(ctx, desiredTower); err != nil {
			return nil, fmt.Errorf("creating Tower: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredTower, nil
	}

	if !equality.Semantic.DeepEqual(desiredTower.Spec, currentTower.Spec) {
		log.V(1).Info("updating", "Tower", name.String())
		desiredTower.Spec = currentTower.Spec
		if err = c.Update(ctx, currentTower); err != nil {
			return nil, fmt.Errorf("updating Role: %w", err)
		}
	}

	return currentTower, nil
}
