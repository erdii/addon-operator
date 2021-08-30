package reconcile

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Namespace reconciles a /v1, Kind=Namespace.
func Namespace(
	ctx context.Context,
	log logr.Logger,
	c client.Client,
	desiredNamespace *corev1.Namespace,
) (
	actualNamespace *corev1.Namespace,
	err error,
) {
	key := client.ObjectKey{
		Name:      desiredNamespace.Name,
		Namespace: desiredNamespace.Namespace,
	}

	// Lookup current version of the object
	actualNamespace = &corev1.Namespace{}
	err = c.Get(ctx, key, actualNamespace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("getting Namespace: %w", err)
	}

	if errors.IsNotFound(err) {
		// Namespace needs to be created
		log.V(1).Info("creating", "Namespace", key.String())
		if err = c.Create(ctx, desiredNamespace); err != nil {
			return nil, fmt.Errorf("creating Namespace: %w", err)
		}
		return desiredNamespace, nil
	}

	return actualNamespace, nil
}
