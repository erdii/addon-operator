package reconcile

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/addon-operator/internal/testutil"
)

func TestNamespace(t *testing.T) {
	t.Run("creates", func(t *testing.T) {
		c := testutil.NewClient()
		ctx := context.Background()
		log := testutil.NewLogger(t)

		sa := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "test-ns",
			},
		}

		notFound := errors.NewNotFound(schema.GroupResource{}, "")
		c.
			On("Get", mock.Anything, client.ObjectKey{
				Name:      "test-sa",
				Namespace: "test-ns",
			}, mock.Anything).
			Return(notFound)

		c.
			On("Create", mock.Anything, mock.Anything, mock.Anything).
			Return(nil)

		_, err := Namespace(ctx, log, c, sa)
		require.NoError(t, err)
		c.AssertCalled(
			t, "Create", mock.Anything, sa, mock.Anything)
	})
}
