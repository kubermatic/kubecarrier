/*
Copyright 2020 The KubeCarrier Authors.

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
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
)

func TimestampProto(t *metav1.Time) (*timestamp.Timestamp, error) {
	if t.IsZero() {
		return nil, nil
	}
	return ptypes.TimestampProto(t.Time)
}

func FromMetav1(obj metav1.ObjectMeta) (*v1.ObjectMeta, error) {
	creationTimestamp, err := TimestampProto(&obj.CreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := TimestampProto(obj.DeletionTimestamp)
	if err != nil {
		return nil, err
	}

	return &v1.ObjectMeta{
		Uid:               string(obj.UID),
		Name:              obj.Name,
		Account:           obj.Namespace,
		CreationTimestamp: creationTimestamp,
		DeletionTimestamp: deletionTimestamp,
		ResourceVersion:   obj.ResourceVersion,
		Labels:            obj.Labels,
		Annotations:       obj.Annotations,
		Generation:        obj.Generation,
	}, nil
}

func ToMetav1(obj *v1.ObjectMeta) (*metav1.ObjectMeta, error) {
	creationTimestamp, err := ptypes.Timestamp(obj.CreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := ptypes.Timestamp(obj.DeletionTimestamp)
	if err != nil {
		return nil, err
	}
	metav1DeletionTimestamp := metav1.NewTime(deletionTimestamp)

	return &metav1.ObjectMeta{
		UID:               types.UID(obj.Uid),
		Name:              obj.Name,
		Namespace:         obj.Account,
		CreationTimestamp: metav1.NewTime(creationTimestamp),
		DeletionTimestamp: &metav1DeletionTimestamp,
		ResourceVersion:   obj.ResourceVersion,
		Labels:            obj.Labels,
		Annotations:       obj.Annotations,
		Generation:        obj.Generation,
	}, nil
}

func FromUnstructured(obj *unstructured.Unstructured) (*v1.ObjectMeta, error) {
	metaCreationTimestamp := obj.GetCreationTimestamp()
	creationTimestamp, err := TimestampProto(&metaCreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := TimestampProto(obj.GetDeletionTimestamp())
	if err != nil {
		return nil, err
	}
	return &v1.ObjectMeta{
		Uid:               string(obj.GetUID()),
		Name:              obj.GetName(),
		Account:           obj.GetNamespace(),
		CreationTimestamp: creationTimestamp,
		DeletionTimestamp: deletionTimestamp,
		ResourceVersion:   obj.GetResourceVersion(),
		Labels:            obj.GetLabels(),
		Annotations:       obj.GetAnnotations(),
		Generation:        obj.GetGeneration(),
	}, nil
}

func SetMetadata(obj *unstructured.Unstructured, metadata *v1.ObjectMeta) error {
	m, err := ToMetav1(metadata)
	if err != nil {
		return err
	}
	obj.SetUID(m.UID)
	obj.SetName(m.Name)
	obj.SetNamespace(m.Namespace)
	obj.SetCreationTimestamp(m.CreationTimestamp)
	obj.SetDeletionTimestamp(m.DeletionTimestamp)
	obj.SetResourceVersion(m.ResourceVersion)
	obj.SetLabels(m.Labels)
	obj.SetAnnotations(m.Annotations)
	obj.SetGeneration(m.Generation)
	return nil
}
