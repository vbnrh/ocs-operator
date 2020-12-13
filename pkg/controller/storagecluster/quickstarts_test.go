package storagecluster

import (
	"context"
	consolev1 "github.com/openshift/api/console/v1"
	api "github.com/openshift/ocs-operator/pkg/apis/ocs/v1"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

const (
	quickstartTestDir = "../../../quickstarts/"
)

var (
	cases = []struct {
		quickstartName string
	}{
		{
			quickstartName: "getting-started-ocs",
		},
		{
			quickstartName: "ocs-configuration",
		},
	}
)

// Test for checking whether the files in the given directory present, readable and parseable into a valid console quickstart object. Does not check validity of its contents.
func TestCheckFileExists(t *testing.T) {
	quickstartDir = quickstartTestDir
	// FetchQuickStart will do the following - check if dir exists, check if file is readable, parse the decoded data into a consolequickstart object and return an array.
	allExpectedQuickStarts, err := FetchQuickStarts(quickstartDir)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(allExpectedQuickStarts))
}
func TestEnsureQuickStarts(t *testing.T) {
	quickstartDir = quickstartTestDir
	cqs := &consolev1.ConsoleQuickStart{}
	reconciler := createFakeStorageClusterReconciler(t, cqs)
	sc := &api.StorageCluster{}
	mockStorageCluster.DeepCopyInto(sc)
	err := reconciler.ensureQuickStarts(sc, reconciler.reqLogger)
	assert.NoError(t, err)
	allExpectedQuickStarts, err := FetchQuickStarts(quickstartDir)
	assert.NoError(t, err)
	for _, c := range cases {
		qs := consolev1.ConsoleQuickStart{}
		err = reconciler.client.Get(context.TODO(), types.NamespacedName{
			Name: c.quickstartName,
		}, &qs)
		assert.NoError(t, err)
		found := consolev1.ConsoleQuickStart{}
		expected := consolev1.ConsoleQuickStart{}
		for _, cqs := range allExpectedQuickStarts {
			if qs.Name == cqs.Name {
				found = qs
				expected = cqs
				break
			}
		}
		assert.Equal(t, expected.Name, found.Name)
		assert.Equal(t, expected.Namespace, found.Namespace)
		assert.Equal(t, expected.Spec.DurationMinutes, found.Spec.DurationMinutes)
		assert.Equal(t, expected.Spec.Introduction, found.Spec.Introduction)
		assert.Equal(t, expected.Spec.DisplayName, found.Spec.DisplayName)
	}
	assert.Equal(t, len(allExpectedQuickStarts), len(getActualQuickStarts(t, cases, &reconciler)))
}

func getActualQuickStarts(t *testing.T, cases []struct {
	quickstartName string
}, reconciler *ReconcileStorageCluster) []consolev1.ConsoleQuickStart {
	allActualQuickStarts := []consolev1.ConsoleQuickStart{}
	for _, c := range cases {
		qs := consolev1.ConsoleQuickStart{}
		err := reconciler.client.Get(context.TODO(), types.NamespacedName{
			Name: c.quickstartName,
		}, &qs)
		if apierrors.IsNotFound(err) {
			continue
		}
		assert.NoError(t, err)
		allActualQuickStarts = append(allActualQuickStarts, qs)
	}
	return allActualQuickStarts
}
