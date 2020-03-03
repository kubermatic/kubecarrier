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

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
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
		return fmt.Errorf("found %d matching conditions, expected 1", len(res))
	}
	rr := res[0]
	if len(rr) != 1 {
		return fmt.Errorf("found %d matching conditions, expected 1", len(rr))
	}
	status := rr[0].String()
	if status != fmt.Sprint(ConditionStatus) {
		return fmt.Errorf("expected condition status %s, got %s", ConditionStatus, status)
	}
	return nil
}

func LogObject(t *testing.T, obj interface{}) {
	t.Helper()
	b, err := json.MarshalIndent(obj, "", "\t")
	require.NoError(t, err)
	t.Log("\n", string(b))
}

func WaitUntilNotFound(ctx context.Context, c *RecordingClient, obj runtime.Object) error {
	c.t.Helper()
	return c.WaitUntilNotFound(ctx, obj)
}

func WaitUntilFound(ctx context.Context, c *RecordingClient, obj runtime.Object) error {
	c.t.Helper()
	return c.WaitUntil(ctx, obj, func(obj runtime.Object, eventType watch.EventType) (b bool, err error) {
		return eventType != watch.Deleted, nil
	})
}

func WaitUntilCondition(ctx context.Context, c *RecordingClient, obj runtime.Object, ConditionType, conditionStatus interface{}) error {
	c.t.Helper()
	err := c.WaitUntil(ctx, obj, func(obj runtime.Object, eventType watch.EventType) (b bool, err error) {
		return ConditionStatusEqual(obj, ConditionType, conditionStatus) == nil, nil
	})

	if err != nil {
		b, marshallErr := json.MarshalIndent(obj, "", "\t")
		if marshallErr != nil {
			return fmt.Errorf("cannot marshall indent obj!!! %v %w", marshallErr, err)
		}
		return fmt.Errorf("%w\n%s", err, string(b))
	}
	return nil
}

func WaitUntilReady(ctx context.Context, c *RecordingClient, obj runtime.Object) error {
	c.t.Helper()
	return WaitUntilCondition(ctx, c, obj, "Ready", "True")
}

func DeleteAndWaitUntilNotFound(ctx context.Context, c *RecordingClient, obj runtime.Object) error {
	c.t.Helper()
	if err := c.Delete(ctx, obj); client.IgnoreNotFound(err) != nil {
		return err
	}
	return WaitUntilNotFound(ctx, c, obj)
}
