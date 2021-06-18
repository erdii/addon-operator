package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	addonsv1alpha1 "github.com/openshift/addon-operator/apis/addons/v1alpha1"
)

// Ensures the presense of an Subscription
func (r *AddonReconciler) ensureSubscription(
	ctx context.Context, log logr.Logger, addon *addonsv1alpha1.Addon) (stop bool, err error) {
	configCtx, err := r.parseAddonInstallConfig(ctx, log, addon)
	if err != nil {
		return false, err
	}
	if configCtx.stop {
		return true, nil
	}

	subscription := &operatorsv1alpha1.Subscription{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addon.Name,
			Namespace: configCtx.targetNamespace,
			Labels:    map[string]string{},
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			CatalogSourceNamespace: configCtx.targetNamespace,
			CatalogSource:          addon.Name,
			Package:                configCtx.packageName,
			Channel:                "stable",
			InstallPlanApproval:    operatorsv1alpha1.ApprovalAutomatic,
		},
	}

	addCommonLabels(subscription.Labels, addon)
	if err := controllerutil.SetControllerReference(addon, subscription, r.Scheme); err != nil {
		return false, fmt.Errorf("setting controller reference: %w", err)
	}

	if err := r.reconcileSubscription(ctx, subscription); err != nil {
		// // write csv/installplan stuff into addon.Status so that next phase can take care of it
		// subscription.Status.InstallPlanRef
	}
	return false, nil
}

// Reconciles the Spec of the given Subscription if needed by updating or creating the Subscription.
func (r *AddonReconciler) reconcileSubscription(
	ctx context.Context, subscription *operatorsv1alpha1.Subscription) error {
	currentSubscription := &operatorsv1alpha1.Subscription{}

	err := r.Get(ctx, client.ObjectKeyFromObject(subscription), currentSubscription)
	if errors.IsNotFound(err) {
		return r.Create(ctx, subscription)
	}
	if err != nil {
		return fmt.Errorf("getting Subscription: %w", err)
	}

	if !equality.Semantic.DeepEqual(currentSubscription.Spec, subscription.Spec) {
		currentSubscription.Spec = subscription.Spec
		return r.Update(ctx, currentSubscription)
	}

	currentSubscription.DeepCopyInto(subscription)
	return nil
}
