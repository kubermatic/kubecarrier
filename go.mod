module github.com/kubermatic/kubecarrier

go 1.14

replace k8s.io/client-go => k8s.io/client-go v0.17.2

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/gernest/wow v0.1.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/flect v0.2.0
	github.com/google/go-cmp v0.3.1
	github.com/huandu/xstrings v1.3.0 // indirect
	github.com/jetstack/cert-manager v0.13.0
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20191206172530-e9b2fee46413
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.5.0
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.1.0
)
