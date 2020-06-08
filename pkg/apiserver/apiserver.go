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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/gorilla/handlers"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8soidc "k8s.io/apiserver/plugin/pkg/authenticator/token/oidc"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	catalogv1alpha1 "github.com/kubermatic/kubecarrier/pkg/apis/catalog/v1alpha1"
	apiserverv1 "github.com/kubermatic/kubecarrier/pkg/apiserver/api/v1"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/auth"
	"github.com/kubermatic/kubecarrier/pkg/apiserver/internal/auth/oidc"
	v1 "github.com/kubermatic/kubecarrier/pkg/apiserver/internal/v1"
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
	address            string
	TLSCertFile        string
	TLSPrivateKeyFile  string
	CORSAllowedOrigins []string
	OIDCOptions        k8soidc.Options
	AuthorizationMode  []string
}

func NewAPIServer() *cobra.Command {
	log := ctrl.Log.WithName("apiserver")
	flags := &flags{}
	cmd := &cobra.Command{
		Args:  cobra.NoArgs,
		Use:   "apiserver",
		Short: "KubeCarrier API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(flags, log)
		},
	}
	cmd.Flags().StringVar(&flags.address, "address", "0.0.0.0:8080", "Address to bind this API server on.")
	cmd.Flags().StringVar(&flags.TLSCertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS. If not provided no TLS security shall be enabled")
	cmd.Flags().StringVar(&flags.TLSPrivateKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching --tls-cert-file.")
	cmd.Flags().StringArrayVar(&flags.CORSAllowedOrigins, "cors-allowed-origins", []string{"*"}, "List of allowed origins for CORS, comma separated. An allowed origin can be a regular expression to support subdomain matching. If this list is empty CORS will not be enabled.")
	cmd.Flags().StringArrayVar(&flags.AuthorizationMode, "authorization-mode", []string{"OIDC"}, "Ordered list of plug-ins to do authorization on secure port. Comma-delimited list of: OIDC,Token")
	oidc.AddOIDCPFlags(&flags.OIDCOptions, cmd.Flags())
	return util.CmdLogMixin(cmd)
}

func runE(flags *flags, log logr.Logger) error {
	// Validation
	if flags.TLSCertFile == "" || flags.TLSPrivateKeyFile == "" {
		return fmt.Errorf("--tls-cert-file or --tls-private-key-file not specified, cannot start")
	}

	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	grpc_zap.ReplaceGrpcLoggerV2(zapLogger)
	var authProviders []auth.AuthProvider

	for _, middleware := range flags.AuthorizationMode {
		switch middleware {
		case "OIDC":
			log.Info("setting up OIDC auth middleware", "iss", flags.OIDCOptions.IssuerURL)
			oidcMiddleware, err := oidc.NewOIDCMiddleware(log, flags.OIDCOptions)
			if err != nil {
				return fmt.Errorf("init OIDC Middleware: %w", err)
			}
			authProviders = append(authProviders, oidcMiddleware)
		default:
			return fmt.Errorf("unknown authorization mode: %v", middleware)
		}
	}

	authFunc := auth.CreateAuthFunction(authProviders)
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_opentracing.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_zap.StreamServerInterceptor(zapLogger),
			grpc_auth.StreamServerInterceptor(authFunc),
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_zap.UnaryServerInterceptor(zapLogger),
			grpc_auth.UnaryServerInterceptor(authFunc),
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	wrappedGrpc := grpcweb.WrapServer(grpcServer)
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
	ctx := context.Background()
	grpcClient, err := createInternalGRPCClient(ctx, flags)
	if err != nil {
		return err
	}
	// Create Kubernetes Client
	cfg := config.GetConfigOrDie()
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return fmt.Errorf("creating rest mapper: %w", err)
	}
	c, err := client.New(cfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("creating dynamic client: %w", err)
	}

	// Set up cache for account
	accountCache, err := cache.New(cfg, cache.Options{
		Scheme: scheme,
		Mapper: mapper,
	})
	if err != nil {
		return fmt.Errorf("creating cache for account: %w", err)
	}
	if err := v1.RegisterAccountUsernameFieldIndex(accountCache); err != nil {
		return fmt.Errorf("fail to register field index for Account Username: %w", err)
	}
	accountClient := &client.DelegatingClient{
		Reader:       accountCache,
		Writer:       c,
		StatusClient: c,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	log.Info("start cache")
	go func() {
		if err := accountCache.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Error(err, "starting cache")
			cancel()
		}
	}()
	if isSynced := accountCache.WaitForCacheSync(ctx.Done()); !isSynced {
		return fmt.Errorf("cache is outdated")
	}

	accountServer := v1.NewAccountServiceServer(accountClient)
	apiserverv1.RegisterAccountServiceServer(grpcServer, accountServer)
	if err := apiserverv1.RegisterAccountServiceHandlerServer(ctx, grpcGatewayMux, accountServer); err != nil {
		return err
	}

	apiserverv1.RegisterKubeCarrierServer(grpcServer, &v1.KubeCarrierServer{})
	if err := apiserverv1.RegisterKubeCarrierHandlerServer(ctx, grpcGatewayMux, &v1.KubeCarrierServer{}); err != nil {
		return err
	}
	offeringServer, err := v1.NewOfferingServiceServer(c, dynamicClient, mapper, scheme)
	if err != nil {
		return err
	}
	apiserverv1.RegisterOfferingServiceServer(grpcServer, offeringServer)
	if err := apiserverv1.RegisterOfferingServiceHandler(ctx, grpcGatewayMux, grpcClient); err != nil {
		return err
	}

	instanceServer := v1.NewInstancesServer(c, mapper)
	apiserverv1.RegisterInstancesServiceServer(grpcServer, instanceServer)
	if err := apiserverv1.RegisterInstancesServiceHandlerServer(ctx, grpcGatewayMux, instanceServer); err != nil {
		return err
	}

	regionServer := v1.NewRegionServiceServer(c)
	apiserverv1.RegisterRegionServiceServer(grpcServer, regionServer)
	if err := apiserverv1.RegisterRegionServiceHandlerServer(ctx, grpcGatewayMux, regionServer); err != nil {
		return err
	}
	providerServer := v1.NewProviderServiceServer(c)
	apiserverv1.RegisterProviderServiceServer(grpcServer, providerServer)
	if err := apiserverv1.RegisterProviderServiceHandlerServer(ctx, grpcGatewayMux, providerServer); err != nil {
		return err
	}

	var handler http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		log.Info("got request for", "path", request.URL.Path)
		if strings.Contains(request.Header.Get("Content-Type"), "application/grpc") {
			wrappedGrpc.ServeHTTP(writer, request)
		} else {
			grpcGatewayMux.ServeHTTP(writer, request)
		}
	})

	if len(flags.CORSAllowedOrigins) > 0 {
		handler = handlers.CORS(
			handlers.AllowedHeaders([]string{
				"X-Requested-With",
				"Content-Type",
				"Authorization",
				"X-grpc-web",
				"X-user-agent",
			}),
			handlers.AllowedMethods([]string{"GET", "POST"}),
			handlers.AllowedOrigins(flags.CORSAllowedOrigins),
		)(handler)
	}

	server := http.Server{
		Handler: handler,
		Addr:    flags.address,
	}

	log.Info("serving serving API-server", "address", flags.address)
	return server.ListenAndServeTLS(flags.TLSCertFile, flags.TLSPrivateKeyFile)
}

func createInternalGRPCClient(ctx context.Context, flags *flags) (*grpc.ClientConn, error) {
	certPool := x509.NewCertPool()
	certs, err := ioutil.ReadFile(flags.TLSCertFile)
	if err != nil {
		return nil, err
	}
	certPool.AppendCertsFromPEM(certs)
	// Create grpc client for watch api
	grpcClient, err := grpc.DialContext(ctx, flags.address, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            certPool,
	})))
	if err != nil {
		return nil, err
	}
	return grpcClient, nil
}
