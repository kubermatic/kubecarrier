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
	fmt "fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

type OfferingGetter interface {
	GetOffering() string
}

func validateOffering(req OfferingGetter) error {
	if req.GetOffering() == "" {
		return fmt.Errorf("missing offering")
	}
	if len(strings.SplitN(req.GetOffering(), ".", 2)) < 2 {
		return fmt.Errorf("offering should have format: {kind}.{apiGroup}")
	}
	return nil
}

type AccountGetter interface {
	GetAccount() string
}

func validateAccount(req AccountGetter) error {
	if req.GetAccount() == "" {
		return fmt.Errorf("missing namespace")
	}
	return nil
}

type NameGetter interface {
	GetName() string
}

func validateName(req NameGetter) error {
	if req.GetName() == "" {
		return fmt.Errorf("missing name")
	}
	return nil
}

type VersionGetter interface {
	GetVersion() string
}

func validateVersion(req VersionGetter) error {
	if req.GetVersion() == "" {
		return fmt.Errorf("missing version")
	}
	return nil
}

type LabelSelectorGetter interface {
	GetLabelSelector() string
}

func validateLabelSelector(req LabelSelectorGetter) error {
	if req.GetLabelSelector() != "" {
		_, err := labels.Parse(req.GetLabelSelector())
		return err
	}
	return nil
}

type LimitGetter interface {
	GetLimit() int64
}

func validateLimit(req LimitGetter) error {
	if req.GetLimit() < 0 {
		return fmt.Errorf("invalid limit: should not be negative number")
	}
	return nil
}

type SpecGetter interface {
	GetSpec() *Instance
}

func validateSpec(req SpecGetter) error {
	if req.GetSpec() == nil {
		return fmt.Errorf("missing spec")
	}
	if req.GetSpec().Metadata == nil && req.GetSpec().Metadata.Name == "" {
		return fmt.Errorf("missing metadata name")
	}
	return nil
}

type ContinueGetter interface {
	GetContinue() string
}

func (req *InstanceGetRequest) Validate() error {
	if err := validateName(req); err != nil {
		return err
	}
	if err := validateOffering(req); err != nil {
		return err
	}
	if err := validateVersion(req); err != nil {
		return err
	}
	if err := validateAccount(req); err != nil {
		return err
	}
	return nil
}

func (req *InstanceDeleteRequest) Validate() error {
	if err := validateName(req); err != nil {
		return err
	}
	if err := validateOffering(req); err != nil {
		return err
	}
	if err := validateVersion(req); err != nil {
		return err
	}
	if err := validateAccount(req); err != nil {
		return err
	}
	return nil
}

func (req *InstanceListRequest) Validate() error {
	if err := validateOffering(req); err != nil {
		return err
	}
	if err := validateVersion(req); err != nil {
		return err
	}
	if err := validateAccount(req); err != nil {
		return err
	}
	if err := validateLimit(req); err != nil {
		return err
	}
	if err := validateLabelSelector(req); err != nil {
		return err
	}
	return nil
}

func (req *InstanceCreateRequest) Validate() error {
	if err := validateOffering(req); err != nil {
		return err
	}
	if err := validateVersion(req); err != nil {
		return err
	}
	if err := validateAccount(req); err != nil {
		return err
	}
	if err := validateSpec(req); err != nil {
		return err
	}
	return nil
}

func (req *AccountListRequest) Validate() error {
	if err := validateLabelSelector(req); err != nil {
		return err
	}
	return nil
}

func (req *GetRequest) Validate() error {
	if err := validateName(req); err != nil {
		return err
	}
	if err := validateAccount(req); err != nil {
		return err
	}
	return nil
}

func (req *ListRequest) Validate() error {
	if err := validateAccount(req); err != nil {
		return err
	}
	if err := validateLabelSelector(req); err != nil {
		return err
	}
	if err := validateLimit(req); err != nil {
		return err
	}
	return nil
}

func (req *WatchRequest) Validate() error {
	if err := validateAccount(req); err != nil {
		return err
	}
	if err := validateLabelSelector(req); err != nil {
		return err
	}
	return nil
}
