module k8c.io/kubecarrier

go 1.14

replace k8s.io/client-go => k8s.io/client-go v0.18.5

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/coreos/go-oidc v2.1.0+incompatible
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/gernest/wow v0.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/flect v0.2.0
	github.com/golang/protobuf v1.3.5
	github.com/google/go-cmp v0.5.0
	github.com/gorilla/handlers v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.14.3
	github.com/improbable-eng/grpc-web v0.12.0
	github.com/jetstack/cert-manager v0.13.0
	github.com/kubermatic/utils v0.0.0-20200724064042-10ba458e0d8d
	github.com/rs/cors v1.7.0 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/tg123/go-htpasswd v1.0.0
	github.com/thetechnick/statik v0.1.8
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	google.golang.org/genproto v0.0.0-20200424135956-bca184e23272
	google.golang.org/grpc v1.28.0
	k8s.io/api v0.18.5
	k8s.io/apiextensions-apiserver v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/apiserver v0.18.5
	k8s.io/cli-runtime v0.18.5
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/component-base v0.18.5
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/krew v0.3.4
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.2.0
)
