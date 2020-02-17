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

package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig = pflag.String("kubeconfig", "", "kubeconfig location")
	as         = pflag.String("as", "", "as which user should the impersonation work")
)

func main() {
	pflag.Parse()
	loader := clientcmd.NewDefaultClientConfigLoadingRules()
	loader.ExplicitPath = *kubeconfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{},
	)
	cfg, err := clientConfig.RawConfig()
	if err != nil {
		panic(err)
	}
	kubeconfigContext := cfg.Contexts[cfg.CurrentContext]
	cfg.AuthInfos[kubeconfigContext.AuthInfo].Impersonate = *as
	if err := clientcmd.WriteToFile(cfg, *kubeconfig); err != nil {
		panic(fmt.Errorf("marshall raw cfg: %w", err))
	}
}
