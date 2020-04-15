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
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
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

type flags struct {
	addr              string
	TLSCertFile       string
	TLSPrivateKeyFile string
	OIDCOptions       oidc.Options
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
	AddPFlags(&flags.OIDCOptions, cmd.Flags())
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

	// v1alpha1 registration
	apiserverv1alpha1.RegisterKubecarrierServer(grpcServer, &kubecarrierHandler{})
	wrapperGRPCServer := grpcweb.WrapServer(grpcServer)
	if err := apiserverv1alpha1.RegisterKubecarrierHandlerServer(context.Background(), grpcGatewayMux, &kubecarrierHandler{}); err != nil {
		return err
	}

	var handler http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Info("got request for", "path", request.URL.Path, "content-type", request.Header.Get("Content-Type"))
		if strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
			wrapperGRPCServer.ServeHTTP(writer, request)
		} else {
			grpcGatewayMux.ServeHTTP(writer, request)
		}
	})

	if flags.OIDCOptions.IssuerURL != "" {
		log.Info("setting up OIDC auth middleware", "iss", flags.OIDCOptions.IssuerURL)
		oidcMiddleware, err := NewOIDCMiddleware(log, flags.OIDCOptions)
		if err != nil {
			return fmt.Errorf("init OIDC Middleware: %w", err)
		}
		handler = oidcMiddleware(handler)
	} else {
		log.Info("skipping OIDC setup")
	}

	server := http.Server{
		Handler: handler,
		Addr:    flags.addr,
	}

	log.Info("booting serving API-server", "addr", flags.addr)
	if flags.TLSCertFile == "" {
		log.V(4).Info("No TLS cert file defined, skipping TLS setup")
		return server.ListenAndServe()
	} else {
		log.Info("using provided TLS cert/key")
		return server.ListenAndServeTLS(flags.TLSCertFile, flags.TLSPrivateKeyFile)
	}
}
