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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

type OptionsRequest interface {
	GetLabelSelector() string
	GetLimit() int64
	GetContinue() string
	GetAccount() string
}

func GetLimitOption(req LimitGetter) (*client.ListOptions, error) {
	if err := validateLimit(req); err != nil {
		return nil, err
	}
	return &client.ListOptions{Limit: req.GetLimit()}, nil
}

func GetContinueOption(req ContinueGetter) (*client.ListOptions, error) {
	return &client.ListOptions{Continue: req.GetContinue()}, nil
}

func GetInNamespaceOption(req AccountGetter) (*client.ListOptions, error) {
	if err := validateAccount(req); err != nil {
		return nil, err
	}
	return &client.ListOptions{Namespace: req.GetAccount()}, nil
}

func GetLabelsSelectorOption(req LabelSelectorGetter) (*client.ListOptions, error) {
	opts := &client.ListOptions{}
	if req.GetLabelSelector() != "" {
		selector, err := labels.Parse(req.GetLabelSelector())
		if err != nil {
			return opts, fmt.Errorf("invalid LabelSelector: %w", err)
		}
		opts.LabelSelector = selector
		return opts, nil
	}
	return opts, nil
}

func (req *ListRequest) GetListOptions() (*client.ListOptions, error) {
	listOptions := &client.ListOptions{}
	// Namespace
	namespace, err := GetInNamespaceOption(req)
	if err != nil {
		return nil, err
	}
	namespace.ApplyToList(listOptions)
	// Limit
	limit, err := GetLimitOption(req)
	if err != nil {
		return nil, err
	}
	limit.ApplyToList(listOptions)
	// Continue
	c, err := GetContinueOption(req)
	if err != nil {
		return nil, err
	}
	c.ApplyToList(listOptions)
	// LabelsSelector
	ls, err := GetLabelsSelectorOption(req)
	if err != nil {
		return nil, err
	}
	ls.ApplyToList(listOptions)
	return listOptions, nil
}

func (req *AccountListRequest) GetListOptions() (*client.ListOptions, error) {
	listOptions := &client.ListOptions{}
	ls, err := GetLabelsSelectorOption(req)
	if err != nil {
		return nil, err
	}
	ls.ApplyToList(listOptions)
	return listOptions, nil
}

func (req *WatchRequest) GetListOptions() (*client.ListOptions, error) {
	listOptions := &client.ListOptions{}
	ls, err := GetLabelsSelectorOption(req)
	if err != nil {
		return nil, err
	}
	ls.ApplyToList(listOptions)
	listOptions.Raw = &metav1.ListOptions{ResourceVersion: req.ResourceVersion}
	return listOptions, nil
}

func (req *InstanceListRequest) GetListOptions() (*client.ListOptions, error) {
	listOptions := &client.ListOptions{}
	// Namespace
	namespace, err := GetInNamespaceOption(req)
	if err != nil {
		return nil, err
	}
	namespace.ApplyToList(listOptions)
	// Limit
	limit, err := GetLimitOption(req)
	if err != nil {
		return nil, err
	}
	limit.ApplyToList(listOptions)
	// Continue
	c, err := GetContinueOption(req)
	if err != nil {
		return nil, err
	}
	c.ApplyToList(listOptions)
	// LabelsSelector
	ls, err := GetLabelsSelectorOption(req)
	if err != nil {
		return nil, err
	}
	ls.ApplyToList(listOptions)
	return listOptions, nil
}
