package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Unstructured reconciles a unstructured.Unstructured object,
// by calling the right typed reconcile function for the given GVK.
// Returns the "real" type, e.g.: *corev1.Service, *appsv1.Deployment.
func Object(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredObject runtime.Object,
) (current metav1.Object, err error) {
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(desiredObject)
	if err != nil {
		return nil, fmt.Errorf("to unstructured converted: %w", err)
	}
	return Unstructured(ctx, log, c, &unstructured.Unstructured{Object: obj})
}
