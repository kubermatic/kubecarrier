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
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

const (
	ProviderNamespaceENV = "PROVIDER_NAMESPACE"
	ServiceClusterENV    = "SERVICE_CLUSTER_NAME"
)

// Task defines single task
type Task struct {
	Name    string
	Program string
	Args    []string
	Env     map[string]string
	LDFlags string
}

func generateIntelijJTasks(tasks []Task, root string) {
	tpl := template.Must(template.New("intelij-task").Funcs(sprig.TxtFuncMap()).Parse(strings.TrimSpace(`
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="kubecarrier:{{.Name}}" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="kubecarrier" />
    <working_directory value="$PROJECT_DIR$/" />
{{- with .LDFlags }}
    <go_parameters value="-i -ldflags &quot;{{ . }}&quot;" />
{{- end }}
{{- with .Args }}
    <parameters value="{{ . | join " " | html }}" />
{{- end }}
{{- with .Env }}
    <envs>
	{{- range $k, $v := . }}
      <env name="{{ $k }}" value="{{ $v }}" />
	{{- end }}
    </envs>
{{- end }}
    <kind value="DIRECTORY" />
    <filePath value="$PROJECT_DIR/|$PROJECT_DIR$/{{ .Program }}" />
    <package value="github.com/kubermatic/kubecarrier" />
    <directory value="$PROJECT_DIR$/{{ .Program }}" />
    <method v="2" />
  </configuration>
</component>
`)))

	err := os.MkdirAll(path.Join(root, ".idea", "runConfigurations"), 0755)
	if err != nil {
		log.Panic(err)
	}

	for _, task := range tasks {
		f, err := os.OpenFile(
			path.Join(root, ".idea", "runConfigurations", "kubecarrier_"+strings.ReplaceAll(task.Name, "/", "__")+".xml"),
			// path.Join(root, ".idea", "runConfigurations", "test.xml"),
			os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
			0755,
		)
		if err != nil {
			log.Panicln(err)
		}
		if err := tpl.Execute(f, task); err != nil {
			log.Panicln(err)
		}
		if err := f.Close(); err != nil {
			log.Panicln(err)
		}
	}
}

type VSCodeLunchConfig struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Request    string            `json:"request"`
	Mode       string            `json:"mode"`
	Program    string            `json:"program"`
	Args       []string          `json:"args,omitempty"`
	Env        map[string]string `json:"env,omitempty"`
	EnvFile    string            `json:"envFile,omitempty"`
	BuildFlags string            `json:"buildFlags,omitempty"`
}

func generateVSCode(tasks []Task, root string) {
	err := os.MkdirAll(path.Join(root, ".vscode"), 0755)
	if err != nil {
		log.Panic(err)
	}
	vscodeLaunchPath := path.Join(root, ".vscode", "launch.json")
	vsCodeConfig := map[string]interface{}{}

	f, err := os.Open(vscodeLaunchPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Panicln("cannot open vscode confgi", err)
	}

	if err == nil {
		if err := json.NewDecoder(f).Decode(&vsCodeConfig); err != nil {
			log.Panicln("cannot decode vsCodeConfig", err)
		}
		if err := f.Close(); err != nil {
			log.Panicln("cannot close f", err)
		}
	}

	_, ok := vsCodeConfig["version"]
	if !ok {
		vsCodeConfig["version"] = "0.2.0"
	}

	vsCodeTasks := make(map[string]VSCodeLunchConfig, len(tasks))
	for _, task := range tasks {
		vsCodeTasks[task.Name] = VSCodeLunchConfig{
			Name:    task.Name,
			Type:    "go",
			Request: "launch",
			Mode:    "auto",
			Program: path.Join("${workspaceFolder}", task.Program),
			Args:    task.Args,
			Env:     task.Env,
			EnvFile: "",
			// for some reason vscode works best with '' but goland with "" surrounding.
			// I'll buy you a beer if you tell me why...
			BuildFlags: "-ldflags '" + task.LDFlags + "'",
		}
	}

	{
		configurations, ok := vsCodeConfig["configurations"]
		if !ok {
			configurations = []interface{}{}
		}
		exitConfigurations := make([]interface{}, 0)
		for _, conf := range configurations.([]interface{}) {
			c := conf.(map[string]interface{})
			task, ok := vsCodeTasks[c["name"].(string)]
			if ok {
				exitConfigurations = append(exitConfigurations, task)
				delete(vsCodeTasks, c["name"].(string))
			} else {
				exitConfigurations = append(exitConfigurations, c)
			}
		}
		for _, task := range vsCodeTasks {
			exitConfigurations = append(exitConfigurations, task)
		}
		vsCodeConfig["configurations"] = exitConfigurations
	}
	b, err := json.MarshalIndent(vsCodeConfig, "", "\t")
	if err != nil {
		log.Panicln("cannot marshal", err)
	}
	if err := ioutil.WriteFile(vscodeLaunchPath, b, 0755); err != nil {
		log.Panicln("cannot write file: ", err)
	}
}

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
	masterKubeconfigPath := path.Join(home, ".kube", "kind-config-kubecarrier-1")
	svcKubeconfigPath := path.Join(home, ".kube", "kind-config-kubecarrier-svc-1")
	flag.Parse()
	fmt.Printf("generating IDE tasks\n")
	fmt.Printf("provider-namespace=%s [use flag or env %s to configure]\n", *providerNamespace, ProviderNamespaceENV)
	fmt.Printf("service-cluster-name=%s [use flag or env %s to configure]\n", *serviceClusterName, ServiceClusterENV)
	var tasks = []Task{
		{
			Name:    "Anchor version",
			Program: "cmd/anchor",
			LDFlags: *ldFlags,
			Args:    []string{"version"},
			Env: map[string]string{
				"KUBECONFIG": masterKubeconfigPath,
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
				"KUBECONFIG": masterKubeconfigPath,
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
				"KUBECONFIG": masterKubeconfigPath,
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
				"KUBECONFIG": masterKubeconfigPath,
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
				"KUBECONFIG": masterKubeconfigPath,
			},
		},
	}

	for _, test := range []string{
		"",
		"AdminSuite",
		"CatapultSuite",
		"FerrySuite",
		"InstallationSuite",
		"ProviderSuite",
		"TenantSuite",
		"VerifySuite",
	} {
		tasks = append(tasks, Task{
			Name:    "e2e:" + test,
			Program: "cmd/anchor",
			LDFlags: *ldFlags,
			Args: []string{
				"e2e-test",
				"run",
				"--test.v",
				"--test.run=" + test,
				"--test-id=1",
			},
		})

	}

	generateVSCode(tasks, ".")
	generateIntelijJTasks(tasks, ".")
}
