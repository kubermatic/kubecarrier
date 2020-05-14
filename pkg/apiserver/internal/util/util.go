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
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TimestampProto(t *metav1.Time) (*timestamp.Timestamp, error) {
	if t.IsZero() {
		return nil, nil
	}
	return ptypes.TimestampProto(t.Time)
}

func ValidateWatchOperation(operation string) error {
	switch operation {
	case "":
	case string(watch.Added):
	case string(watch.Modified):
	case string(watch.Deleted):
	case string(watch.Bookmark):
	case string(watch.Error):
	default:
		return fmt.Errorf("invalid watch operation type: %s", operation)
	}
	return nil
}
