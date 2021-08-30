module github.com/openshift/addon-operator

go 1.16

require (
	github.com/containers/image/v5 v5.16.0
	github.com/go-logr/logr v0.4.0
	github.com/openshift/hive/apis v0.0.0-20210823153512-6cfe97689937
	github.com/operator-framework/api v0.8.1
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.20.6
	k8s.io/apiextensions-apiserver v0.20.1
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.20.6
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/yaml v1.2.0
)
