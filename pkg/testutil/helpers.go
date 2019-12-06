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

package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/jsonpath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ConditionStatusEqual(obj runtime.Object, ConditionType, ConditionStatus interface{}) error {
	jp := jsonpath.New("condition")
	if err := jp.Parse(fmt.Sprintf(`{.status.conditions[?(@.type=="%s")].status}`, ConditionType)); err != nil {
		return err
	}
	res, err := jp.FindResults(obj)
	if err != nil {
		return fmt.Errorf("cannot find results: %w", err)
	}
	if len(res) != 1 {
		return fmt.Errorf("found %d values, expected 1", len(res))
	}
	rr := res[0]
	if len(rr) != 1 {
		return fmt.Errorf("found %d values, expected 1", len(rr))
	}
	status := rr[0].String()
	if status != fmt.Sprint(ConditionStatus) {
		return fmt.Errorf("expected %s, got %s", ConditionStatus, status)
	}
	return nil
}

func LogObject(t *testing.T, obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "\t")
	require.NoError(t, err)
	t.Log("\n", string(b))
}

func WaitUntilNotFound(c client.Client, obj runtime.Object) error {
	o, ok := obj.(metav1.Object)
	if !ok {
		return fmt.Errorf("%T does not implement metav1.Object", obj)
	}

	return wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		err = c.Get(context.Background(), types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		}, obj)
		switch {
		case errors.IsNotFound(err):
			return true, nil
		case err == nil:
			return false, nil
		default:
			return false, err

		}
	})
}

func WaitUntilCondition(c client.Client, obj runtime.Object, ConditionType, ConditionStatus interface{}) error {
	o, ok := obj.(metav1.Object)
	if !ok {
		return fmt.Errorf("%T does not implement metav1.Object", obj)
	}
	var lastErr error
	err := wait.Poll(time.Second, 30*time.Second, func() (done bool, err error) {
		err = c.Get(context.Background(), types.NamespacedName{
			Namespace: o.GetNamespace(),
			Name:      o.GetName(),
		}, obj)
		switch {
		case errors.IsNotFound(err):
			return false, nil
		case err == nil:
			lastErr = ConditionStatusEqual(obj, ConditionType, ConditionStatus)
			return lastErr == nil, nil
		default:
			return false, err
		}
	})

	if err != nil {
		if lastErr != nil {
			return lastErr
		}
		return err
	}
	return nil
}
