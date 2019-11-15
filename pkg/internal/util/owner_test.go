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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubermatic/kubecarrier/pkg/internal/testutil"
)

func TestOwnerReverseFieldIndex(t *testing.T) {
	t.Parallel()
	env := envtest.Environment{}
	cfg, err := env.Start()
	require.NoError(t, err, "cannot start test env")
	defer func() {
		require.NoError(t, env.Stop())
	}()

	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: "0",
	})
	require.NoError(t, err, "cannot create manager")
	log := testutil.NewLogger(t)

	log.Info("Adding indices")

	// make sure there's no errors
	assert.NoError(t, AddOwnerReverseFieldIndex(mgr.GetFieldIndexer(), log, &corev1.Pod{}), "adding cm reverse index")
	assert.NoError(t, AddOwnerReverseFieldIndex(mgr.GetFieldIndexer(), log, &corev1.ConfigMap{}), "adding configmap reverse index")

	// make sure it errors out on duplicate object type
	assert.Error(t, AddOwnerReverseFieldIndex(mgr.GetFieldIndexer(), log, &corev1.Pod{}), "adding cm reverse index")

	cl := mgr.GetClient()
	configMaps := make([]*corev1.ConfigMap, 5)
	for i := range configMaps {
		configMaps[i] = &corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("cm-%d", i),
				Namespace: "default",
			},
		}
	}

	ownerA := configMaps[3]
	ownerAFilter, err := OwnedBy(ownerA, mgr.GetScheme())
	require.NoError(t, err, "converting to ownerReference")

	ownerB := configMaps[4]
	ownerBFilter, err := OwnedBy(ownerB, mgr.GetScheme())
	require.NoError(t, err, "converting to ownerReference")

	extractErr := func(changed bool, err error) error { return err }

	require.NoError(t, extractErr(InsertOwnerReference(ownerA, configMaps[0], mgr.GetScheme())), "inserting owner reference")
	require.NoError(t, extractErr(InsertOwnerReference(ownerA, configMaps[1], mgr.GetScheme())), "inserting owner reference")
	require.NoError(t, extractErr(InsertOwnerReference(ownerB, configMaps[1], mgr.GetScheme())), "inserting owner reference")
	require.NoError(t, extractErr(InsertOwnerReference(ownerB, configMaps[2], mgr.GetScheme())), "inserting owner reference")

	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go func() {
		require.NoError(t, mgr.Start(ctx.Done()))
	}()

	for _, cm := range configMaps {
		require.NoError(t, cl.Create(ctx, cm), "creating configMaps")
	}

	configMapList := corev1.ConfigMapList{}
	listItems := func(lst corev1.ConfigMapList) (items []string) {
		for _, it := range lst.Items {
			items = append(items, it.Name)
		}
		return
	}

	// wait for field indices finishing their work
	mgr.GetCache().WaitForCacheSync(ctx.Done())
	// and a bit more for good measure
	time.Sleep(10 * time.Millisecond)

	t.Log("listing ownerA owned configmaps")
	require.NoError(t, cl.List(ctx, &configMapList, client.InNamespace("default"), ownerAFilter), "listing configMaps")
	t.Logf("ownerA got owned items: %v", listItems(configMapList))
	assert.Len(t, configMapList.Items, 2)
	for _, it := range configMapList.Items {
		assert.Contains(t, []string{"cm-0", "cm-1"}, it.Name, "wrong selection")
	}

	t.Log("listing ownerB owned configmaps")
	require.NoError(t, cl.List(ctx, &configMapList, client.InNamespace("default"), ownerBFilter), "listing configMaps")
	t.Logf("ownerB got owned items: %v", listItems(configMapList))
	assert.Len(t, configMapList.Items, 2)
	for _, it := range configMapList.Items {
		assert.Contains(t, []string{"cm-1", "cm-2"}, it.Name, "wrong selection")
	}
}

func TestCRUDOwnerMethods(t *testing.T) {
	sc := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(sc))

	ownerA := &corev1.Pod{ObjectMeta: v1.ObjectMeta{
		Name:      "A",
		Namespace: "default",
	}}
	ownerB := &corev1.Pod{ObjectMeta: v1.ObjectMeta{
		Name:      "B",
		Namespace: "default",
	}}

	obj := &corev1.Pod{ObjectMeta: v1.ObjectMeta{
		Name:      "obj",
		Namespace: "default",
	}}

	t.Log("===== Initial state =====")
	for i, own := range []*corev1.Pod{ownerA, ownerB} {
		t.Logf("===== Adding owner %s =====", own.Name)
		check := func() {
			refs, err := getRefs(obj)
			require.NoError(t, err, "getRefs")
			assert.Len(t, refs, i+1, "getRefs")
		}
		changed, err := InsertOwnerReference(own, obj, sc)
		require.NoError(t, err, "insert owner reference")
		assert.True(t, changed, "obj changed status")
		check()

		t.Log("======== idempotent repeat =====")
		changed, err = InsertOwnerReference(own, obj, sc)
		require.NoError(t, err, "insert owner reference")
		assert.False(t, changed)
		check()
	}
}
