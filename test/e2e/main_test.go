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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"sigs.k8s.io/kind/pkg/cluster"
)

var (
	// Current test id. The created clusters shall be prefixed accordingly
	testID string

	// reuse existing test environment, if exists
	reuse bool

	// Keep the existing test clusters after the test finishes
	keep bool

	masterKubeconfig          string
	masterExternalKubeconfig  string
	serviceKubeconfig         string
	serviceExternalKubecongif string
)

func RandomID() string {
	return strconv.FormatUint(rand.Uint64(), 36)[:4]
}

func TestMain(m *testing.M) {

	defaultTestID := os.ExpandEnv("BUILD_ID")
	if defaultTestID == "" {
		defaultTestID = RandomID()
	}

	flag.StringVar(
		&testID,
		"test-id",
		defaultTestID,
		"test id to use",
	)
	flag.BoolVar(
		&reuse,
		"reuse",
		true,
		"Reuse existing test environment if exists",
	)
	flag.BoolVar(
		&keep,
		"keep",
		true,
		"Keep existing test clusters after tests finishes",
	)
	flag.Parse()

	p := cluster.NewProvider()
	clusters, err := p.List()
	if err != nil {
		log.Panic(err)
	}
	fmt.Print(clusters)
}
