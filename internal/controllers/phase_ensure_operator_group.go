package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

// Ensures the presence or absence of an OperatorGroup depending on the Addon install type.
func (r *AddonReconciler) ensureOperatorGroup(
	ctx context.Context, log logr.Logger, addon *addonsv1alpha1.Addon) (stop bool, err error) {
	configCtx, err := r.parseAddonInstallConfig(ctx, log, addon)
	if err != nil {
		return false, err
	}
	if configCtx.stop {
		return true, nil
	}

	desiredOperatorGroup := &operatorsv1.OperatorGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon.Name,
			Namespace: configCtx.targetNamespace,
			Labels:    map[string]string{},
		},
	}
	if addon.Spec.Install.Type == addonsv1alpha1.OwnNamespace {
		desiredOperatorGroup.Spec.TargetNamespaces = []string{configCtx.targetNamespace}
	}

	addCommonLabels(desiredOperatorGroup.Labels, addon)
	if err := controllerutil.SetControllerReference(addon, desiredOperatorGroup, r.Scheme); err != nil {
		return false, fmt.Errorf("setting controller reference: %w", err)
	}

	return false, r.reconcileOperatorGroup(ctx, desiredOperatorGroup)
}

// Reconciles the Spec of the given OperatorGroup if needed by updating or creating the OperatorGroup.
func (r *AddonReconciler) reconcileOperatorGroup(
	ctx context.Context, operatorGroup *operatorsv1.OperatorGroup) error {
	currentOperatorGroup := &operatorsv1.OperatorGroup{}

	err := r.Get(ctx, client.ObjectKeyFromObject(operatorGroup), currentOperatorGroup)
	if errors.IsNotFound(err) {
		return r.Create(ctx, operatorGroup)
	}
	if err != nil {
		return fmt.Errorf("getting OperatorGroup: %w", err)
	}

	if !equality.Semantic.DeepEqual(currentOperatorGroup.Spec, operatorGroup.Spec) {
		return r.Update(ctx, operatorGroup)
	}
	return nil
}
