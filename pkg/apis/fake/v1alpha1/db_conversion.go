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

package v1alpha1

import (
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/conversion"

	v1 "github.com/kubermatic/kubecarrier/pkg/apis/fake/v1"
)

var _ conversion.Convertible = (*DB)(nil)

func (src *DB) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1.DB)

	conn := src.Status.Connection
	if conn != nil {
		dst.Status.Connection = fmt.Sprintf("%s:%s:%s", conn.Endpoint, conn.Username, conn.Name)
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.DatabaseName = src.Spec.DatabaseName
	dst.Spec.RootPassword = src.Spec.RootPassword
	dst.Spec.DatabaseUser = src.Spec.DatabaseUser
	dst.Spec.DatabasePassword = src.Spec.DatabasePassword
	dst.Spec.Config = v1.Config{
		ReadyAfterSeconds:    src.Spec.Config.ReadyAfterSeconds,
		DeletionAfterSeconds: src.Spec.Config.DeletionAfterSeconds,
		Create:               v1.OperationFlagType(string(src.Spec.Config.Create)),
		Update:               v1.OperationFlagType(src.Spec.Config.Update),
		Delete:               v1.OperationFlagType(src.Spec.Config.Delete),
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Phase = v1.DBPhaseType(string(src.Status.Phase))
	for _, cond := range src.Status.Conditions {
		dst.Status.Conditions = append(dst.Status.Conditions, v1.DBCondition{
			Type:               v1.DBConditionType(string(cond.Type)),
			Status:             v1.ConditionStatus(string(cond.Status)),
			LastTransitionTime: cond.LastTransitionTime,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})

	}

	return nil
}
func (dst *DB) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1.DB)

	conn := src.Status.Connection
	if conn != "" {
		connParts := strings.Split(conn, ":")
		if len(connParts) != 3 {
			return fmt.Errorf("invalid connection: not a standard 3-field connection string")
		}
		dst.Status.Connection = &Connection{
			Endpoint: connParts[0],
			Username: connParts[1],
			Name:     connParts[2],
		}
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.DatabaseName = src.Spec.DatabaseName
	dst.Spec.RootPassword = src.Spec.RootPassword
	dst.Spec.DatabaseUser = src.Spec.DatabaseUser
	dst.Spec.DatabasePassword = src.Spec.DatabasePassword
	dst.Spec.Config = Config{
		ReadyAfterSeconds:    src.Spec.Config.ReadyAfterSeconds,
		DeletionAfterSeconds: src.Spec.Config.DeletionAfterSeconds,
		Create:               OperationFlagType(string(src.Spec.Config.Create)),
		Update:               OperationFlagType(src.Spec.Config.Update),
		Delete:               OperationFlagType(src.Spec.Config.Delete),
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Phase = DBPhaseType(string(src.Status.Phase))
	for _, cond := range src.Status.Conditions {
		dst.Status.Conditions = append(dst.Status.Conditions, DBCondition{
			Type:               DBConditionType(string(cond.Type)),
			Status:             ConditionStatus(string(cond.Status)),
			LastTransitionTime: cond.LastTransitionTime,
			Reason:             cond.Reason,
			Message:            cond.Message,
		})

	}
	return nil
}
