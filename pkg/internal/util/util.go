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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// object generic k8s object with metav1 and runtime Object interfaces implemented
type object interface {
	runtime.Object
	metav1.Object
}

type ConditionPatch struct {
	Condition interface{}
	Schema    *runtime.Scheme
}

func (c *ConditionPatch) Type() types.PatchType {
	return types.ApplyPatchType
}

func (c *ConditionPatch) Data(obj runtime.Object) ([]byte, error) {
	u := &unstructured.Unstructured{}
	gvk, err := apiutil.GVKForObject(obj, c.Schema)
	if err != nil {
		return nil, err
	}
	u.SetGroupVersionKind(gvk)
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}
	u.SetNamespace(accessor.GetNamespace())
	u.SetName(accessor.GetName())
	u.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			c.Condition,
		},
	}
	return json.Marshal(u)
}

var _ client.Patch = (*ConditionPatch)(nil)
