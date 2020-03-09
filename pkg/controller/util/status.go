package util

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	ocsv1 "github.com/openshift/ocs-operator/pkg/apis/ocs/v1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	corev1 "k8s.io/api/core/v1"
)

// These constants represent the overall Phase as used by .Status.Phase
var (
	// PhaseIgnored is used when a resource is ignored
	PhaseIgnored = "Ignored"
	// PhaseProgressing is used when SetProgressingCondition is called
	PhaseProgressing = "Progressing"
	// PhaseError is used when SetErrorCondition is called
	PhaseError = "Error"
	// PhaseReady is used when SetCompleteCondition is called
	PhaseReady = "Ready"
	// PhaseNotReady is used when waiting for system to be ready
	// after reconcile is successful
	PhaseNotReady = "Not Ready"
	// PhaseClusterExpanding is used when cluster is expanding capacity
	PhaseClusterExpanding = "Expanding Capacity"
	// PhaseDeleting is used when cluster is deleting
	PhaseDeleting = "Deleting"
	// PhaseConnecting is used when cluster is connecting to external cluster
	PhaseConnecting = "Connecting"
	//PhaseConnected is used when cluster has connected to external cluster
	PhaseConnected = "Connected"
)

// SetProgressingCondition sets the ProgressingCondition to True and other conditions to
// false or Unknown. Used when we are just starting to reconcile, and there are no existing
// conditions.
func SetProgressingCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    ocsv1.ConditionReconcileComplete,
		Status:  corev1.ConditionUnknown,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionAvailable,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionProgressing,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionDegraded,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionUpgradeable,
		Status:  corev1.ConditionUnknown,
		Reason:  reason,
		Message: message,
	})
}

// SetErrorCondition sets the ConditionReconcileComplete to False in case of any errors
// during the reconciliation process.
func SetErrorCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    ocsv1.ConditionReconcileComplete,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
}

// SetCompleteCondition sets the ConditionReconcileComplete to True and other Conditions
// to indicate that the reconciliation process has completed successfully.
func SetCompleteCondition(conditions *[]conditionsv1.Condition, reason string, message string) {
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    ocsv1.ConditionReconcileComplete,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionAvailable,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionProgressing,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionDegraded,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionUpgradeable,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
}

// MapCephClusterNegativeConditions maps the status states from CephCluster resource into ocs status conditions.
// This will only look for negative conditions: !Available, Degraded, Progressing
func MapCephClusterNegativeConditions(conditions *[]conditionsv1.Condition, found *cephv1.CephCluster) {
	switch found.Status.State {
	case cephv1.ClusterStateCreating:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionProgressing,
			Status:  corev1.ConditionTrue,
			Reason:  "ClusterStateCreating",
			Message: fmt.Sprintf("CephCluster is creating: %v", string(found.Status.Message)),
		})
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionUpgradeable,
			Status:  corev1.ConditionFalse,
			Reason:  "ClusterStateCreating",
			Message: fmt.Sprintf("CephCluster is creating: %v", string(found.Status.Message)),
		})
	case cephv1.ClusterStateUpdating:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionProgressing,
			Status:  corev1.ConditionTrue,
			Reason:  "ClusterStateUpdating",
			Message: fmt.Sprintf("CephCluster is updating: %v", string(found.Status.Message)),
		})
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionUpgradeable,
			Status:  corev1.ConditionFalse,
			Reason:  "ClusterStateUpdating",
			Message: fmt.Sprintf("CephCluster is updating: %v", string(found.Status.Message)),
		})
	case cephv1.ClusterStateError:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionAvailable,
			Status:  corev1.ConditionFalse,
			Reason:  "ClusterStateError",
			Message: fmt.Sprintf("CephCluster error: %v", string(found.Status.Message)),
		})
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionDegraded,
			Status:  corev1.ConditionTrue,
			Reason:  "ClusterStateError",
			Message: fmt.Sprintf("CephCluster error: %v", string(found.Status.Message)),
		})
	}
}


// MapExternalCephClusterNegativeConditions maps the status states from CephCluster resource into ocs status conditions.
// This will only look for negative conditions: !Available, Degraded, Progressing
func MapExternalCephClusterNegativeConditions(conditions *[]conditionsv1.Condition, found *cephv1.CephCluster) {
	switch found.Status.State {
	case cephv1.ClusterStateConnecting:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    ocsv1.ConditionExternalClusterConnecting,
			Status:  corev1.ConditionTrue,
			Reason:  "ExternalClusterStateConnecting",
			Message: fmt.Sprintf("ExternalCephCluster is trying to connect: %v", found.Status.Message),
		})
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    ocsv1.ConditionExternalClusterConnected,
			Status:  corev1.ConditionFalse,
			Reason:  "ExternalClusterStateConnecting",
			Message: fmt.Sprintf("ExternalCephCluster is trying to connect: %v", found.Status.Message),
		})
	case cephv1.ClusterStateError:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    ocsv1.ConditionExternalClusterConnected,
			Status:  corev1.ConditionFalse,
			Reason:  "ExternalClusterStateError",
			Message: fmt.Sprintf("External CephCluster error: %v", string(found.Status.Message)),
		})
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    ocsv1.ConditionExternalClusterConnecting,
			Status:  corev1.ConditionFalse,
			Reason:  "ExternalClusterStateError",
			Message: fmt.Sprintf("External CephCluster error: %v", string(found.Status.Message)),
		})
	default:
		conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionDegraded,
			Status:  corev1.ConditionTrue,
			Reason:  "ExternalClusterStateUnknownCondition",
			Message: fmt.Sprintf("External CephCluster Unknown Condition: %v", string(found.Status.Message)),
		})
	}
}

// MapCephClusterNoConditions sets status conditions to progressing. Used when component operator isn't
// reporting any status, and we have to assume progress.
func MapCephClusterNoConditions(conditions *[]conditionsv1.Condition, reason string, message string) {
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionAvailable,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionProgressing,
		Status:  corev1.ConditionTrue,
		Reason:  reason,
		Message: message,
	})
	conditionsv1.SetStatusCondition(conditions, conditionsv1.Condition{
		Type:    conditionsv1.ConditionUpgradeable,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
}

// won't override a status condition of the same type and status
func setStatusConditionIfNotPresent(conditions *[]conditionsv1.Condition, condition conditionsv1.Condition) {

	foundCondition := conditionsv1.FindStatusCondition(*conditions, condition.Type)
	if foundCondition != nil && foundCondition.Status == condition.Status {
		// already exists
		return
	}

	conditionsv1.SetStatusCondition(conditions, condition)
}

// MapNoobaaNegativeConditions records noobaa related conditions
// This will only look for negative conditions: !Available, Degraded, Progressing
func MapNoobaaNegativeConditions(conditions *[]conditionsv1.Condition, found *nbv1.NooBaa) {

	if found == nil {
		setStatusConditionIfNotPresent(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionDegraded,
			Status:  corev1.ConditionTrue,
			Reason:  "NoobaaNotFound",
			Message: fmt.Sprintf("Waiting on Nooba instance creation"),
		})
		return
	}

	switch found.Status.Phase {
	case nbv1.SystemPhaseRejected:
		setStatusConditionIfNotPresent(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionDegraded,
			Status:  corev1.ConditionTrue,
			Reason:  "NoobaaSpecRejected",
			Message: fmt.Sprintf("Noobaa object's configuration is rejected by the noobaa operator"),
		})
	case "", nbv1.SystemPhaseVerifying, nbv1.SystemPhaseCreating, nbv1.SystemPhaseConnecting, nbv1.SystemPhaseConfiguring:
		setStatusConditionIfNotPresent(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionProgressing,
			Status:  corev1.ConditionTrue,
			Reason:  "NoobaaInitializing",
			Message: fmt.Sprintf("Waiting on Nooba instance to finish initialization"),
		})
	case nbv1.SystemPhaseReady:
		// no-op. Ready isn't a negative case
	default:
		setStatusConditionIfNotPresent(conditions, conditionsv1.Condition{
			Type:    conditionsv1.ConditionDegraded,
			Status:  corev1.ConditionTrue,
			Reason:  "NoobaaPhaseUnknown",
			Message: fmt.Sprintf("Noobaa phase %s is unknown", found.Status.Phase),
		})
	}

}
