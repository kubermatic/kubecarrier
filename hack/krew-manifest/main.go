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
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/krew/pkg/constants"
	"sigs.k8s.io/krew/pkg/index"
	"sigs.k8s.io/yaml"
)

func main() {
	//https://github.com/goreleaser/goreleaser/issues/1375
	version := flag.String("version", "v0.1.0", "kubecarrier version")
	flag.Parse()

	kubecarrier := &index.Plugin{
		Spec: index.PluginSpec{
			Version:          *version,
			ShortDescription: "KubeCarrier installation management and control",
			Description: strings.TrimSpace(`
The kubecarrier plugin for installing, upgrading and performing management tasks in the Kubecarrier system.

KubeCarrier is an open source system for managing applications and services across multiple Kubernetes Clusters; providing a framework to centralize the management of services and provide these services with external users in a self service catalog.,

Overview:
	https://github.com/kubermatic/kubecarrier/
Installation:
	https://github.com/kubermatic/kubecarrier/#install-kubecarrier
Quick Start:
	https://github.com/kubermatic/kubecarrier/blob/master/README.md
`),
			Caveats: strings.TrimSpace(`
The kubecarrier is a management plugin for controlling the KubeCarrier system. See https://github.com/kubermatic/kubecarrier/ for more details.

For quick start installation run:

kubectl kubecarrier setup

For all other commands, options and how to use the system read the official documentation.
`),
			Homepage:  "https://github.com/kubermatic/kubecarrier",
			Platforms: nil,
		},
	}
	kubecarrier.Kind = constants.PluginKind
	kubecarrier.APIVersion = constants.CurrentAPIVersion
	kubecarrier.Name = "kubecarrier"

	for _, arch := range []string{"amd64", "386"} {
		for _, os := range []struct {
			osName     string
			binaryExt  string
			archiveExt string
		}{
			{
				osName:     "linux",
				binaryExt:  "",
				archiveExt: ".tar.gz",
			},
			{
				osName:     "darwin",
				binaryExt:  "",
				archiveExt: ".tar.gz",
			},
			{
				osName:     "windows",
				binaryExt:  ".exe",
				archiveExt: ".zip",
			},
		} {
			platform := index.Platform{
				URI: fmt.Sprintf("https://github.com/kubermatic/kubecarrier/releases/download/%s/kubecarrier_%s_%s%s", *version, os.osName, arch, os.archiveExt),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"os":   os.osName,
						"arch": arch,
					},
				},
				Bin: "kubecarrier" + os.binaryExt,
			}
			pp, err := filepath.Abs(path.Join("dist", fmt.Sprintf("kubecarrier_%s_%s%s", os.osName, arch, os.archiveExt)))
			if err != nil {
				panic(err)
			}
			b, err := ioutil.ReadFile(pp)
			if err != nil {
				panic(err)
			}
			sum := sha256.Sum256(b)
			platform.Sha256 = hex.EncodeToString(sum[:])
			kubecarrier.Spec.Platforms = append(kubecarrier.Spec.Platforms, platform)
		}
	}

	b, err := yaml.Marshal(kubecarrier)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))
}
