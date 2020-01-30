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

package webhook

import (
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

const DNS1123LabelDescription = "A DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character. (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?'"

var dns1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// IsDNS1123Label validates if string s is a validated DNS 1123 label.
// A DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character.
func IsDNS1123Label(s string) bool {
	return dns1123LabelRegex.MatchString(s)
}

// GenerateMutateWebhookPath and GenerateValidatingWebhookPath are used to generate the Path to register webhooks for runtime.Object.
// They are similar to the functions in the controller-runtime package:
// https://github.com/kubernetes-sigs/controller-runtime/blob/dc8357113a904bf02721efcde5d92937be39031c/pkg/builder/webhook.go#L158-L166

func GenerateMutateWebhookPath(obj runtime.Object, scheme *runtime.Scheme) (string, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return "", err
	}
	return "/mutate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind), nil
}

func GenerateValidateWebhookPath(obj runtime.Object, scheme *runtime.Scheme) (string, error) {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return "", err
	}
	return "/validate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind), nil
}
