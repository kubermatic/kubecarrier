/*
Copyright 2019 The Kubecarrier Authors.

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

package e2e

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/create"
)

var (
	// Current e2e-test id. The created clusters shall be prefixed accordingly
	testID string

	// reuse existing e2e-test environment, if exists
	reuse bool

	// Keep the existing e2e-test clusters after the e2e-test finishes
	keep bool

	masterKubeconfig          []byte
	masterExternalKubeconfig  []byte
	serviceKubeconfig         []byte
	serviceExternalKubecongif []byte
)

func RandomID() string {
	return strconv.FormatUint(rand.Uint64(), 36)[:4]
}

func TestMain(m *testing.M) {
	defaultTestID := os.ExpandEnv("${BUILD_ID}")
	if defaultTestID == "" {
		defaultTestID = RandomID()
	}

	flag.StringVar(
		&testID,
		"e2e-test-id",
		defaultTestID,
		"e2e-test id to use",
	)
	flag.BoolVar(
		&reuse,
		"reuse",
		true,
		"Reuse existing e2e-test environment if exists",
	)
	flag.BoolVar(
		&keep,
		"keep",
		true,
		"Keep existing e2e-test clusters after tests finishes",
	)
	flag.Parse()

	for _, cl := range []struct {
		Name               string
		internalKubeconfig []byte
		externalKubeconfig []byte
	}{
		{
			Name:               "kubecarrier-" + testID,
			internalKubeconfig: masterKubeconfig,
			externalKubeconfig: masterExternalKubeconfig,
		},
		{
			Name:               "kubecarrier-svc-" + testID,
			internalKubeconfig: serviceKubeconfig,
			externalKubeconfig: serviceExternalKubecongif,
		},
	} {
		know, err := cluster.IsKnown(cl.Name)
		if err != nil {
			log.Panic(err)
		}
		ctx := cluster.NewContext(cl.Name)
		switch {
		case !know:
			log.Printf("creating cluster %s", ctx.Name())
			if err := ctx.Create(
				// To apply the defaults...
				create.WithConfigFile(""),
				create.Retain(keep),
				create.WithNodeImage(""),
				create.WaitForReady(time.Duration(0)),
			); err != nil {
				log.Panic(err)
			}
		case reuse && know:
			log.Printf("found existing kind cluster %s, reusing it\n", ctx.Name())
		case !reuse && know:
			log.Printf("found existing kind cluster %s, but reuse is disabled\n", ctx.Name())
			os.Exit(1)
		default:
		}

		if !keep {
			defer func() {
				log.Printf("deleting cluster %s", ctx.Name())
				if err := ctx.Delete(); err != nil {
					log.Printf("cannot delete cluster %s", ctx.Name())
				}

			}()
		}

		log.Printf("for cluster %s kubeconifg is at %s", ctx.Name(), ctx.KubeConfigPath())

		cl.externalKubeconfig, err = ioutil.ReadFile(ctx.KubeConfigPath())
		if err != nil {
			log.Panicf("cannot read kubeconfig: %v", err)
		}

		log.Printf("master internal Kubeconfig %v", string(masterKubeconfig))
		log.Printf("master external Kubeconfig %v", string(masterExternalKubeconfig))

		// List nodes by cluster context name
		n, err := ctx.ListInternalNodes()
		if err != nil {
			log.Panicf("cannot list internal nodes: %v", err)
		}

		cmdNode := n[0].Command("cat", "/etc/kubernetes/admin.conf")
		b := new(bytes.Buffer)
		cmdNode.SetStdout(b)
		if err := cmdNode.Run(); err != nil {
			log.Panicf("cannot read the internal kubeconfig %v", err)
		}
		cl.internalKubeconfig = b.Bytes()
	}

}
