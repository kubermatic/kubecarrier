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
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/structured-merge-diff/v3/typed"
	"sigs.k8s.io/yaml"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
)

func SplitStatusFields(fields []catalogv1alpha1.FieldPath) (
	statusFields []catalogv1alpha1.FieldPath,
	otherFields []catalogv1alpha1.FieldPath,
) {
	for _, field := range fields {
		path := strings.Trim(field.JSONPath, ".")
		if path == "status" || strings.HasPrefix(path, "status.") {
			statusFields = append(statusFields, field)
			continue
		}

		otherFields = append(otherFields, field)
	}
	return
}

func CopyFields(
	src, dest *unstructured.Unstructured,
	fields []catalogv1alpha1.FieldPath) error {

	for _, field := range fields {
		path := strings.Trim(field.JSONPath, ".") // trim trailing and leading dots
		fields := strings.Split(path, ".")
		value, exists, err := unstructured.NestedFieldCopy(src.Object, fields...)
		if err != nil {
			return fmt.Errorf("lookup path in %s: %w", src.GetKind(), err)
		}

		if !exists {
			continue
		}

		if err := unstructured.SetNestedField(dest.Object, value, fields...); err != nil {
			return fmt.Errorf("update path in %s: %w", dest.GetKind(), err)
		}
	}
	return nil
}

func VersionExposeConfigForVersion(
	exposeConfigs []catalogv1alpha1.VersionExposeConfig, version string,
) (catalogv1alpha1.VersionExposeConfig, bool) {
	for _, exposeConfig := range exposeConfigs {
		for _, exposeVersion := range exposeConfig.Versions {
			if exposeVersion == version {
				return exposeConfig, true
			}
		}
	}
	return catalogv1alpha1.VersionExposeConfig{}, false
}

func FormPatch(patchField *runtime.RawExtension) (patch map[string]interface{}, err error) {
	if patchField != nil {
		patch = make(map[string]interface{})
		if err := yaml.Unmarshal(patchField.Raw, &patch); err != nil {
			return nil, fmt.Errorf("patch isn")
		}
	}
	return
}

func BuildProviderObj(tenantObj *unstructured.Unstructured, providerObj *unstructured.Unstructured, scheme *runtime.Scheme, elevateFields []catalogv1alpha1.FieldPath, patch interface{}) error {
	providerObj.SetName(tenantObj.GetName())
	providerObj.SetNamespace(tenantObj.GetNamespace())

	// controller reference without UID is invalid!
	// it will fail upon creation
	if tenantObj.GetUID() != "" {
		err := controllerutil.SetControllerReference(
			tenantObj, providerObj, scheme)
		if err != nil {
			return fmt.Errorf("set controller reference: %w", err)
		}
	}

	if err := CopyFields(tenantObj, providerObj, elevateFields); err != nil {
		return fmt.Errorf("copy fields: %w", err)
	}

	if patch != nil {
		lhs, err := typed.DeducedParseableType.FromUnstructured(providerObj.Object)
		if err != nil {
			return fmt.Errorf("cannot convert to patch: %w", err)
		}
		patch, err := typed.DeducedParseableType.FromUnstructured(patch)
		if err != nil {
			return fmt.Errorf("cannot convert to patch: %w", err)
		}
		val, err := lhs.Merge(patch)
		if err != nil {
			return err
		}
		providerObj.Object = val.AsValue().Unstructured().(map[string]interface{})
	}
	return nil
}
