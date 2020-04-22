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

package apiserver

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

type flags struct {
	port int
}

func NewAPIServer() *cobra.Command {
	log := ctrl.Log.WithName("api-server")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "api-server",
		Short: "KubeCarrier API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}
	cmd.Flags().IntVar(&flags.port, "port", 8080, "port to serve this API server at")
	return util.CmdLogMixin(cmd)
}

func runE(flags *flags, log logr.Logger) error {
	log.Info("booting serving API-server", "port", flags.port)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(writer, err.Error())
			return
		}
		log.Info("new request", "body", string(body), "method", request.Method, "host", request.Host, "path", request.URL.Path)
		fmt.Fprintln(writer, "hello world!")
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", flags.port), nil)
}
