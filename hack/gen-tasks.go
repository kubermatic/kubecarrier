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
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/kubermatic/kubecarrier/pkg/ide"
)

const (
	ProviderNamespaceENV = "PROVIDER_NAMESPACE"
	ServiceClusterENV    = "SERVICE_CLUSTER_NAME"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panicln("cannot find user's home dir", err)
	}
	ldFlags := flag.String("ldflags", "", "ld-flags for go's binaries")
	providerNamespace := flag.String(
		"provider-namespace",
		os.ExpandEnv("$"+ProviderNamespaceENV),
		"provider namespace to use in task generation",
	)
	serviceClusterName := flag.String(
		"service-cluster-name",
		os.ExpandEnv("$"+ServiceClusterENV),
		"provider namespace to use in task generation",
	)
	managementKubeconfigPath := path.Join(home, ".kube", "kind-config-kubecarrier-1")
	svcKubeconfigPath := path.Join(home, ".kube", "kind-config-kubecarrier-svc-1")
	flag.Parse()
	fmt.Printf("generating IDE tasks\n")
	fmt.Printf("provider-namespace=%s [use flag or env %s to configure]\n", *providerNamespace, ProviderNamespaceENV)
	fmt.Printf("service-cluster-name=%s [use flag or env %s to configure]\n", *serviceClusterName, ServiceClusterENV)
	var tasks = []ide.Task{
		{
			Name:    "API Server",
			Program: "cmd/api-server",
			LDFlags: *ldFlags,
			Args: []string{
				"--address=:8090",
				"--oidc-issuer-url=https://accounts.google.com",
				"--oidc-client-id=640570493642-p10tov8pbr0b5kplri6to2fumbkrf397.apps.googleusercontent.com",
				"--oidc-username-claim=email",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Kubecarrier version",
			Program: "cmd/kubectl-kubecarrier",
			LDFlags: *ldFlags,
			Args:    []string{"version"},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Manager",
			Program: "cmd/manager",
			LDFlags: *ldFlags,
			Args: []string{
				"--kubecarrier-system-namespace=kubecarrier-system",
				"--enable-leader-election=false",
				"--metrics-addr=0",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Operator",
			Program: "cmd/operator",
			LDFlags: *ldFlags,
			Args: []string{
				"--enable-leader-election=false",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Ferry",
			Program: "cmd/ferry",
			LDFlags: *ldFlags,
			Args: []string{
				"--provider-namespace=" + *providerNamespace,
				"--service-cluster-name=" + *serviceClusterName,
				"--service-cluster-kubeconfig=" + svcKubeconfigPath,
				"--enable-leader-election=false",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Catapult",
			Program: "cmd/catapult",
			LDFlags: *ldFlags,
			Args: []string{
				"--provider-namespace=" + *providerNamespace,
				"--service-cluster-name=" + *serviceClusterName,
				"--service-cluster-kubeconfig=" + svcKubeconfigPath,
				"--enable-leader-election=false",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
		{
			Name:    "Elevator",
			Program: "cmd/elevator",
			LDFlags: *ldFlags,
			Args: []string{
				"--provider-namespace=default",
			},
			Env: map[string]string{
				"KUBECONFIG": managementKubeconfigPath,
			},
		},
	}

	for _, test := range []string{
		"",
		"InstallationSuite",
		"Integration",
		"Integration/account",
		"Integration/api-server",
		"Integration/catalog",
		"Integration/cli",
		"Integration/derivedCR",
		"Integration/serviceCluster",
		"Scenarios",
		"Scenarios/simple",
		"VerifySuite",
	} {
		tasks = append(tasks, ide.Task{
			Name:    "e2e:" + test,
			Program: "cmd/kubectl-kubecarrier",
			LDFlags: *ldFlags,
			Args: []string{
				"e2e-test",
				"run",
				"--test.v",
				"--test.run=" + test,
				"--test-id=1",
				"--clean-up-strategy=always",
			},
		})

	}

	ide.GenerateVSCode(tasks, ".")
	ide.GenerateIntelijJTasks(tasks, ".")
}
