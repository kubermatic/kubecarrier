module github.com/kubermatic/kubecarrier

go 1.13

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a

require (
	github.com/gernest/wow v0.1.0
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.3.1
	github.com/rakyll/statik v0.1.6
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.3.0
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.1.0
)
