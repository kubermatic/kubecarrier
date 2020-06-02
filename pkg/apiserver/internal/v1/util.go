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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/util"
)

const (
	accountUserFieldIndex = "account.kubecarrier.io/user"
)

// RegisterAccountUsernameFieldIndex adds a field index for user names in Account.Spec.Subjects.
func RegisterAccountUsernameFieldIndex(indexer client.FieldIndexer) error {
	return indexer.IndexField(
		&catalogv1alpha1.Account{}, accountUserFieldIndex,
		func(obj runtime.Object) (values []string) {
			account := obj.(*catalogv1alpha1.Account)
			for _, subject := range account.Spec.Subjects {
				values = append(values, subject.Name)
			}
			return
		})
}

func accountByUsernameListOption(username string) client.ListOption {
	return client.MatchingFields{
		accountUserFieldIndex: username,
	}
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
