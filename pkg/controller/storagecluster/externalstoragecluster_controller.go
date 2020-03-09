package storagecluster

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	objectreferencesv1 "github.com/openshift/custom-resource-status/objectreferences/v1"
	ocsv1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1"
	statusutil "github.com/openshift/ocs-operator/pkg/controller/util"
	"github.com/operator-framework/operator-sdk/pkg/ready"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReconcileExternalStorageCluster reconciles the external Storage Cluster
func (r *ReconcileStorageCluster) ReconcileExternalStorageCluster(sc *ocsv1.StorageCluster) (reconcile.Result, error) {
	reqLogger := r.reqLogger.WithValues("Request.Namespace", sc.Namespace, "Request.Name", sc.Name)
	reqLogger.Info("Reconciling External StorageCluster")
	var err error

	// Add conditions if there are none
	if sc.Status.Conditions == nil {
		reason := ocsv1.ReconcileInit
		message := "Initializing External StorageCluster"
		statusutil.SetProgressingCondition(&sc.Status.Conditions, reason, message)
		err = r.client.Status().Update(context.TODO(), sc)
		if err != nil {
			reqLogger.Error(err, "Failed to add conditions to status")
			return reconcile.Result{}, err
		}
	}

	// Check GetDeletionTimestamp to determine if the object is under deletion
	if sc.GetDeletionTimestamp().IsZero() {
		if !contains(sc.GetFinalizers(), storageClusterFinalizer) {
			reqLogger.Info("Finalizer not found for storagecluster. Adding finalizer")
			sc.ObjectMeta.Finalizers = append(sc.ObjectMeta.Finalizers, storageClusterFinalizer)
			if err := r.client.Update(context.TODO(), sc); err != nil {
				reqLogger.Error(err, "Failed to update storagecluster with finalizer")
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is marked for deletion
		sc.Status.Phase = statusutil.PhaseDeleting
		phaseErr := r.client.Status().Update(context.TODO(), sc)
		if phaseErr != nil {
			reqLogger.Error(phaseErr, "Failed to set PhaseDeleting")
		}
		if contains(sc.GetFinalizers(), storageClusterFinalizer) {
			isDeleted, err := r.deleteResources(sc, reqLogger)
			if err != nil {
				// If the dependencies failed to delete because of errors, retry again
				return reconcile.Result{}, err
			}
			if isDeleted {
				reqLogger.Info("Removing finalizer")
				// Once all finalizers have been removed, the object will be deleted
				sc.ObjectMeta.Finalizers = remove(sc.ObjectMeta.Finalizers, storageClusterFinalizer)
				if err := r.client.Update(context.TODO(), sc); err != nil {
					reqLogger.Error(err, "Failed to remove finalizer from storagecluster")
					return reconcile.Result{}, err
				}
			} else {
				// Watch resources and events and reconcile.
				return reconcile.Result{}, nil
			}
		}
		reqLogger.Info("Object is terminated, skipping reconciliation")
		return reconcile.Result{}, nil
	}

	externalCephCluster := newExternalCephCluster(sc, r.cephImage)
	// Set StorageCluster instance as the owner and controller
	if err := controllerutil.SetControllerReference(sc, externalCephCluster, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this CephCluster already exists
	found := &cephv1.CephCluster{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: externalCephCluster.Name, Namespace: externalCephCluster.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Creating External CephCluster")
			if err := r.client.Create(context.TODO(), externalCephCluster); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, err
	}
	// Update the CephCluster if it is not in the desired state
	if !reflect.DeepEqual(externalCephCluster.Spec, found.Spec) {
		reqLogger.Info("Updating spec for External CephCluster")
		sc.Status.Phase = string(found.Status.State)
		err = r.client.Status().Update(context.TODO(), sc)
		if err != nil {
			reqLogger.Error(err, "Failed to add conditions to status")
			return reconcile.Result{}, err
		}
		found.Spec = externalCephCluster.Spec
		if err := r.client.Update(context.TODO(), found); err != nil {
			return reconcile.Result{}, err
		}
	}

	// in-memory conditions should start off empty. It will only ever hold
	// negative conditions (!Available, Degraded, Progressing)
	r.conditions = nil
	// Start with empty r.phase
	r.phase = ""

	for _, f := range []func(*ocsv1.StorageCluster, logr.Logger) error{
		// Add support for additional resources here
		r.ensureCephConfig,
		r.ensureExternalCephCluster,
		r.ensureNoobaaSystem,
	} {
		err = f(sc, reqLogger)
		if r.phase == statusutil.PhaseClusterExpanding {
			sc.Status.Phase = statusutil.PhaseClusterExpanding
			phaseErr := r.client.Status().Update(context.TODO(), sc)
			if phaseErr != nil {
				reqLogger.Error(phaseErr, "Failed to set PhaseClusterExpanding")
			}
		} else {
			if sc.Status.Phase != statusutil.PhaseReady && sc.Status.Phase != statusutil.PhaseConnecting && sc.Status.Phase != statusutil.PhaseConnected {
				sc.Status.Phase = statusutil.PhaseProgressing
				phaseErr := r.client.Status().Update(context.TODO(), sc)
				if phaseErr != nil {
					reqLogger.Error(phaseErr, "Failed to set PhaseProgressing")
				}
			}
		}
		if err != nil {
			reason := ocsv1.ReconcileFailed
			message := fmt.Sprintf("Error while reconciling: %v", err)
			statusutil.SetErrorCondition(&sc.Status.Conditions, reason, message)
			sc.Status.Phase = statusutil.PhaseError
			// don't want to overwrite the actual reconcile failure
			uErr := r.client.Status().Update(context.TODO(), sc)
			if uErr != nil {
				reqLogger.Error(uErr, "Failed to update status")
			}
			return reconcile.Result{}, err
		}
	}

	// All component operators are in a happy state.
	if r.conditions == nil {
		reqLogger.Info("No component operator reported negatively")
		reason := ocsv1.ReconcileCompleted
		message := ocsv1.ReconcileCompletedMessage
		statusutil.SetCompleteCondition(&sc.Status.Conditions, reason, message)

		// If no operator whose conditions we are watching reports an error, then it is safe
		// to set readiness.
		r := ready.NewFileReady()
		err = r.Set()
		if err != nil {
			reqLogger.Error(err, "Failed to mark operator ready")
			return reconcile.Result{}, err
		}
		if sc.Status.Phase != statusutil.PhaseClusterExpanding {
			sc.Status.Phase = statusutil.PhaseReady
		}
	} else {
		// If any component operator reports negatively we want to write that to
		// the instance while preserving it's lastTransitionTime.
		// For example, consider the resource has the Available condition
		// type with type "False". When reconciling the resource we would
		// add it to the in-memory representation of OCS's conditions (r.conditions)
		// and here we are simply writing it back to the server.
		// One shortcoming is that only one failure of a particular condition can be
		// captured at one time (ie. if resource1 and resource2 are both reporting !Available,
		// you will only see resource2q as it updates last).
		for _, condition := range r.conditions {
			conditionsv1.SetStatusCondition(&sc.Status.Conditions, condition)
		}
		reason := ocsv1.ReconcileCompleted
		message := ocsv1.ReconcileCompletedMessage
		conditionsv1.SetStatusCondition(&sc.Status.Conditions, conditionsv1.Condition{
			Type:    ocsv1.ConditionReconcileComplete,
			Status:  corev1.ConditionTrue,
			Reason:  reason,
			Message: message,
		})

		// If for any reason we marked ourselves !upgradeable...then unset readiness
		if conditionsv1.IsStatusConditionFalse(sc.Status.Conditions, conditionsv1.ConditionUpgradeable) {
			r := ready.NewFileReady()
			err = r.Unset()
			if err != nil {
				reqLogger.Error(err, "Failed to mark operator unready")
				return reconcile.Result{}, err
			}
		}
		if sc.Status.Phase != statusutil.PhaseClusterExpanding {
			if conditionsv1.IsStatusConditionTrue(sc.Status.Conditions, conditionsv1.ConditionProgressing) {
				sc.Status.Phase = statusutil.PhaseProgressing
			} else if conditionsv1.IsStatusConditionFalse(sc.Status.Conditions, conditionsv1.ConditionUpgradeable) {
				sc.Status.Phase = statusutil.PhaseNotReady
			} else {
				sc.Status.Phase = statusutil.PhaseError
			}
		}
	}
	if phaseErr := r.client.Status().Update(context.TODO(), sc); phaseErr != nil {
		reqLogger.Error(phaseErr, "Failed to update status")
		return reconcile.Result{}, phaseErr
	}

	return reconcile.Result{}, nil
}

func newExternalCephCluster(sc *ocsv1.StorageCluster, cephImage string) *cephv1.CephCluster {
	labels := map[string]string{
		"app": sc.Name,
	}
	externalCephCluster := &cephv1.CephCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generateNameForExternalCephCluster(sc),
			Namespace: sc.Namespace,
			Labels:    labels,
		},
		Spec: cephv1.ClusterSpec{
			External: cephv1.ExternalSpec{
				Enable: true,
			},
			DataDirHostPath: "/var/lib/rook",
			CephVersion: cephv1.CephVersionSpec{
				Image: cephImage,
			},
		},
	}
	return externalCephCluster
}

func (r *ReconcileStorageCluster) ensureExternalCephCluster(
	sc *ocsv1.StorageCluster, reqLogger logr.Logger) error {

	// Define a new CephCluster object
	cephCluster := newExternalCephCluster(sc, r.cephImage)

	// Set StorageCluster instance as the owner and controller
	if err := controllerutil.SetControllerReference(sc, cephCluster, r.scheme); err != nil {
		return err
	}

	// Check if this CephCluster already exists
	found := &cephv1.CephCluster{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cephCluster.Name, Namespace: cephCluster.Namespace}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("Creating External CephCluster: v", r.cephImage)
			return r.client.Create(context.TODO(), cephCluster)
		}
		return err
	}

	// Update the CephCluster if it is not in the desired state
	if !reflect.DeepEqual(cephCluster.Spec, found.Spec) {
		reqLogger.Info("Updating spec for External CephCluster")
		sc.Status.Phase = string(found.Status.State)
		if err := r.client.Status().Update(context.TODO(), sc); err != nil {
			reqLogger.Error(err, "Failed to add conditions to status")
			return err
		}
		found.Spec = cephCluster.Spec
		return r.client.Update(context.TODO(), found)
	}

	// Add it to the list of RelatedObjects if found
	objectRef, err := reference.GetReference(r.scheme, found)
	if err != nil {
		return err
	}
	objectreferencesv1.SetObjectReference(&sc.Status.RelatedObjects, *objectRef)

	// Handle CephCluster resource status
	if found.Status.State == "" {
		reqLogger.Info("CephCluster resource is not reporting status.")
		// What does this mean to OCS status? Assuming progress.
		reason := "CephClusterStatus"
		message := "CephCluster resource is not reporting status"
		statusutil.MapCephClusterNoConditions(&r.conditions, reason, message)
	} else {
		// Interpret CephCluster status and set any negative conditions
		// here negative conditions for external cluster has tobe set
		statusutil.MapExternalCephClusterNegativeConditions(&r.conditions,found)
	}

	// When phase is expanding, wait for CephCluster state to be updating
	// this means expansion is in progress and overall system is progressing
	// else expansion is not yet triggered
	if sc.Status.Phase == statusutil.PhaseClusterExpanding &&
		found.Status.State != cephv1.ClusterStateUpdating {
		r.phase = statusutil.PhaseClusterExpanding
	}

	if found.Status.State == cephv1.ClusterStateConnecting {
		sc.Status.Phase = statusutil.PhaseConnecting
	}

	if found.Status.State == cephv1.ClusterStateConnected {
		sc.Status.Phase = statusutil.PhaseConnected
	}

	return nil
}
