module github.com/kubermatic/kubecarrier

go 1.13

require (
	github.com/google/go-cmp v0.3.1
	github.com/spf13/cobra v0.0.3
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.9.1
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.1
	sigs.k8s.io/kind v0.5.1
)
