package testutil

import (
	"context"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

// custom testify/mock matchers
var (
	// core
	IsCoreV1NamespacePtr     = mock.IsType(&corev1.Namespace{})
	IsCoreV1NamespaceListPtr = mock.IsType(&corev1.NamespaceList{})

	// olm
	IsOperatorsV1Alpha1CatalogSourcePtr = mock.IsType(&operatorsv1alpha1.CatalogSource{})

	// addon.managed.openshift.io/v1alpha1
	IsAddonsv1alpha1AddonPtr             = mock.IsType(&addonsv1alpha1.Addon{})
	IsAddonsv1alpha1AddonListPtr         = mock.IsType(&addonsv1alpha1.AddonList{})
	IsAddonsv1alpha1AddonOperatorPtr     = mock.IsType(&addonsv1alpha1.AddonOperator{})
	IsAddonsv1alpha1AddonOperatorListPtr = mock.IsType(&addonsv1alpha1.AddonOperatorList{})

	// misc
	IsContext   = mock.IsType(context.TODO())
	IsObjectKey = mock.IsType(client.ObjectKey{})
)
