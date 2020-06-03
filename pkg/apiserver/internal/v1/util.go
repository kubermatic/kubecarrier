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

package v1

import (
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

func ToMetav1(obj *v1.ObjectMeta) (*metav1.ObjectMeta, error) {

	objMeta := &metav1.ObjectMeta{
		UID:             types.UID(obj.Uid),
		Name:            obj.Name,
		Namespace:       obj.Account,
		ResourceVersion: obj.ResourceVersion,
		Labels:          obj.Labels,
		Annotations:     obj.Annotations,
		Generation:      obj.Generation,
	}
	if obj.CreationTimestamp != nil {
		timestamp, err := ptypes.Timestamp(obj.CreationTimestamp)
		if err != nil {
			return nil, err
		}
		objMeta.CreationTimestamp = metav1.NewTime(timestamp)
	}
	if obj.DeletionTimestamp != nil {
		timestamp, err := ptypes.Timestamp(obj.DeletionTimestamp)
		if err != nil {
			return nil, err
		}
		metav1DeletionTimestamp := metav1.NewTime(timestamp)
		objMeta.DeletionTimestamp = &metav1DeletionTimestamp
	}
	return objMeta, nil
}

func FromUnstructured(obj *unstructured.Unstructured) (*v1.ObjectMeta, error) {
	metaCreationTimestamp := obj.GetCreationTimestamp()
	creationTimestamp, err := util.TimestampProto(&metaCreationTimestamp)
	if err != nil {
		return nil, err
	}
	deletionTimestamp, err := util.TimestampProto(obj.GetDeletionTimestamp())
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

func convertImage(in *catalogv1alpha1.Image) (out *v1.Image) {
	return &v1.Image{
		MediaType: in.MediaType,
		Data:      in.Data,
	}
}

func convertObjectMeta(in metav1.ObjectMeta) (out *v1.ObjectMeta, err error) {
	creationTimestamp, err := util.TimestampProto(&in.CreationTimestamp)
	if err != nil {
		return out, err
	}
	deletionTimestamp, err := util.TimestampProto(in.DeletionTimestamp)
	if err != nil {
		return out, err
	}
	out = &v1.ObjectMeta{
		Uid:               string(in.UID),
		Name:              in.Name,
		Account:           in.Namespace,
		CreationTimestamp: creationTimestamp,
		DeletionTimestamp: deletionTimestamp,
		ResourceVersion:   in.ResourceVersion,
		Labels:            in.Labels,
		Annotations:       in.Annotations,
		Generation:        in.Generation,
	}
	return
}

func convertListMeta(in metav1.ListMeta) (out *v1.ListMeta) {
	out = &v1.ListMeta{
		Continue:        in.Continue,
		ResourceVersion: in.ResourceVersion,
	}
	return
}

type toGRPCStatus interface {
	GRPCStatus() *status.Status
}
