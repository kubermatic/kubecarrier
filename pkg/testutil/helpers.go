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
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func AssertConditionStatus(t *testing.T, obj runtime.Object, ConditionType interface{}, ConditionStatus interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	status := reflect.ValueOf(obj).Elem().FieldByName("Status").Addr()
	method := status.MethodByName("GetCondition")
	require.False(t, method.IsZero())
	require.Equal(t, 1, method.Type().NumIn())
	require.Equal(t, 2, method.Type().NumOut())
	res := method.Call([]reflect.Value{reflect.ValueOf(ConditionType)})

	cond, ok := res[0], res[1]
	if !assert.True(t, ok.Bool(), "condition not found") {
		return
	}
	require.Equal(t, cond.FieldByName("Status").Type().String(), reflect.ValueOf(ConditionStatus).Type().String(), "wrong status type")
	if fmt.Sprint(ConditionStatus) != cond.FieldByName("Status").String() {
		assert.Fail(t,
			fmt.Sprintf(
				"condition %s: expected %s, got %s", reflect.ValueOf(ConditionType).Type().String(), ConditionStatus, cond.FieldByName("Status"),
			), msgAndArgs...)
	}
}

func LogObject(t *testing.T, obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "\t")
	require.NoError(t, err)
	t.Log("\n", string(b))
}
