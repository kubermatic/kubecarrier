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

package checkers

import (
	"bytes"
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

// certManager checks if the cert-manager related deployments are ready.
type certManagerChecker struct {
	client          client.Client
	certManagerCRDs []string
}

func (c *certManagerChecker) check() error {
	ctx := context.Background()
	var errBuffer bytes.Buffer

	for _, crdName := range c.certManagerCRDs {
		crd := &apiextensionsv1.CustomResourceDefinition{}
		if err := c.client.Get(ctx, types.NamespacedName{
			Name: crdName,
		}, crd); err != nil {
			errBuffer.WriteString(err.Error() + "\n")
		} else {
			if !util.CRDIsEstablished(crd) {
				errBuffer.WriteString(fmt.Sprintf("crd %s is not established\n", crd.Name))
			}
		}
	}
	if errBuffer.Len() > 0 {
		return fmt.Errorf(errBuffer.String())
	}
	return nil
}

func (c *certManagerChecker) name() string {
	return "CertManager"
}
