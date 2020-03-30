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
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	grpcgatewayruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	apiserverv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/apiserver/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

const (
	componentTower = "Tower"
)

type flags struct {
	port int
}

func NewAPIServer() *cobra.Command {
	log := ctrl.Log.WithName("api-server")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   componentTower,
		Short: "KubeCarrier Tower",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}
	cmd.Flags().IntVar(&flags.port, "port", 8080, "port to serve this API server at")
	return util.CmdLogMixin(cmd)
}

func runE(flags *flags, log logr.Logger) error {
	log.Info("booting serving API-server", "port", flags.port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", flags.port))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	grpcGatewayMux := grpcgatewayruntime.NewServeMux()
	apiserverv1alpha1.RegisterVersionServiceServer(grpcServer, &versionHandler{})
	err = apiserverv1alpha1.RegisterVersionServiceHandlerServer(context.Background(), grpcGatewayMux, &versionHandler{})
	if err != nil {
		return err
	}

	router := mux.NewRouter()
	var v1alpha1 http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
		log.Info("got request for", "path", request.URL.Path)
		if strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(writer, request)
		} else {
			grpcGatewayMux.ServeHTTP(writer, request)
		}
	}
	router.PathPrefix("/v1alpha1").Handler(http.StripPrefix("/v1alpha1", v1alpha1))
	return http.Serve(lis, router)
}
