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
	"encoding/json"
	"errors"
	fmt "fmt"

	"sigs.k8s.io/yaml"
)

const (
	EncodingJSON = "json"
	EncodingYAML = "yaml"
)

func NewRawObject(format string, data []byte) (*RawObject, error) {
	if format != EncodingJSON && format != EncodingYAML {
		return nil, errors.New("unsupported format")
	}
	return &RawObject{Encoding: format, Data: data}, nil
}

func NewJSONRawObject(data []byte) *RawObject {
	ro, _ := NewRawObject(EncodingJSON, data)
	return ro
}

func NewYAMLRawObject(data []byte) *RawObject {
	ro, _ := NewRawObject(EncodingYAML, data)
	return ro
}

func (ro *RawObject) Unmarshal(v interface{}) error {
	switch ro.Encoding {
	case EncodingJSON:
		return json.Unmarshal(ro.Data, v)
	case EncodingYAML:
		return yaml.Unmarshal(ro.Data, v)
	default:
		return fmt.Errorf("unsupported encoding format: %s", ro.Encoding)
	}
}
