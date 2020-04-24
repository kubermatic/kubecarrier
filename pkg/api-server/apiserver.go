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
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	apiv1 "github.com/kubermatic/kubecarrier/pkg/api/v1"
	apiservicev1 "github.com/kubermatic/kubecarrier/pkg/api/v1/services"
	apiserverv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/apiserver/v1alpha1"
	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	"github.com/kubermatic/kubecarrier/pkg/internal/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(catalogv1alpha1.AddToScheme(scheme))
}

type flags struct {
	addr              string
	TLSCertFile       string
	TLSPrivateKeyFile string
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
	cmd.Flags().StringVar(&flags.addr, "addr", ":8080", "port to serve this API server at")
	cmd.Flags().StringVar(&flags.TLSCertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. If not provided no TLS security shall be enabled")
	cmd.Flags().StringVar(&flags.TLSPrivateKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	return util.CmdLogMixin(cmd)
}

func runE(flags *flags, log logr.Logger) error {
	grpcServer := grpc.NewServer()
	grpcGatewayMux := gwruntime.NewServeMux(
		gwruntime.WithProtoErrorHandler(func(ctx context.Context, serveMux *gwruntime.ServeMux, marshaler gwruntime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			const fallback = `{"error": "failed to marshal error message"}`
			writer.Header().Del("Trailer")
			writer.Header().Set("Content-Type", marshaler.ContentType())
			s, ok := status.FromError(err)
			if !ok {
				s = status.New(codes.Unknown, err.Error())
			}
			buf, marshalerr := marshaler.Marshal(s.Proto())
			if marshalerr != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				_, _ = io.WriteString(writer, fallback)
				return
			}

			st := gwruntime.HTTPStatusFromCode(s.Code())
			writer.WriteHeader(st)
			_, _ = writer.Write(buf)
		}),
	)

	// Create Kubernetes Client
	cfg := config.GetConfigOrDie()
	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return fmt.Errorf("creating rest mapper: %w", err)
	}
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})

	// v1alpha1 registration
	apiserverv1alpha1.RegisterKubecarrierServer(grpcServer, &kubecarrierHandler{})
	if err := apiserverv1alpha1.RegisterKubecarrierHandlerServer(context.Background(), grpcGatewayMux, &kubecarrierHandler{}); err != nil {
		return err
	}
	offeringServer := apiservicev1.NewOfferingServiceServer(c)
	apiv1.RegisterOfferingServiceServer(grpcServer, offeringServer)
	if err := apiv1.RegisterOfferingServiceHandlerServer(context.Background(), grpcGatewayMux, offeringServer); err != nil {
		return err
	}

	var handlerFunc http.HandlerFunc = func(writer http.ResponseWriter, request *http.Request) {
		log.Info("got request for", "path", request.URL.Path)
		if strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(writer, request)
		} else {
			grpcGatewayMux.ServeHTTP(writer, request)
		}
	}

	server := http.Server{
		Handler: handlerFunc,
		Addr:    flags.addr,
	}

	log.Info("booting serving API-server", "addr", flags.addr)
	if flags.TLSCertFile == "" {
		log.Info("No TLS cert file defined, skipping TLS setup")
		return server.ListenAndServe()
	} else {
		log.Info("using provided TLS cert/key")
		return server.ListenAndServeTLS(flags.TLSCertFile, flags.TLSPrivateKeyFile)
	}
}
