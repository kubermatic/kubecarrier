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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Deployment reconciles a apps/v1, Kind=Deployment.
func Deployment(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredDeployment *appsv1.Deployment,
) (currentDeployment *appsv1.Deployment, err error) {
	name := types.NamespacedName{
		Name:      desiredDeployment.Name,
		Namespace: desiredDeployment.Namespace,
	}

	// Lookup current version of the object
	currentDeployment = &appsv1.Deployment{}
	err = c.Get(ctx, name, currentDeployment)
	if err != nil && !errors.IsNotFound(err) {
		// unexpected error
		return nil, fmt.Errorf("getting Deployment: %w", err)
	}

	// Keep the replicas resource cleaned
	// telepresence scales the original deployment to 0 replicas before starting a new one
	if currentDeployment.Spec.Replicas != nil {
		desiredDeployment.Spec.Replicas = currentDeployment.Spec.Replicas
	}

	if errors.IsNotFound(err) {
		// Deployment needs to be created
		log.V(1).Info("creating", "Deployment", name.String())
		if err = c.Create(ctx, desiredDeployment); err != nil {
			return nil, fmt.Errorf("creating Deployment: %w", err)
		}
		// no need to check for updates, object was created just now
		return desiredDeployment, nil
	}

	if !equality.Semantic.DeepEqual(desiredDeployment.Spec, currentDeployment.Spec) {
		// desired and current Deployment .Spec are not equal -> trigger an update
		log.V(1).Info("updating", "Deployment", name.String())
		currentDeployment.Spec = desiredDeployment.Spec
		if err = c.Update(ctx, currentDeployment); err != nil {
			return nil, fmt.Errorf("updating Deployment: %w", err)
		}
	}

	return currentDeployment, nil
}
