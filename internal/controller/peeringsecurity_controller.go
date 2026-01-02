/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/forge"
	"github.com/riccardotornesello/liqo-security-manager/internal/controller/utils"
)

// PeeringSecurityReconciler reconciles a PeeringSecurity object
type PeeringSecurityReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	// Condition Types
	ConditionTypeReady = "Ready"

	// Reasons
	ReasonClusterIDError   = "ClusterIDExtractionFailed"
	ReasonFabricSyncFailed = "FabricSyncFailed"
	ReasonFabricSynced     = "FabricSynced"

	// Event Types (Normal vs Warning is managed by k8s, here we define the reasons for events)
	EventReasonReconcileError = "ReconcileError"
	EventReasonSynced         = "Synced"
)

// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=security.liqo.io,resources=peeringsecurities/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PeeringSecurityReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO: make sure the cluster exists
	// TODO: handle the case of multiple PeeringSecurity in the same cluster

	logger := log.FromContext(ctx)

	// Retrieve the PeeringSecurity resource
	cfg := &securityv1.PeeringSecurity{}
	if err := r.Client.Get(ctx, req.NamespacedName, cfg); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("missing PeeringSecurity resource, skipping reconciliation")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to get the PeeringSecurity %q: %w", req.NamespacedName, err)
	}

	logger.Info("reconciling PeeringSecurity")

	// Extract Cluster ID from Namespace
	clusterID, err := utils.ExtractClusterID(req.Namespace)
	if err != nil {
		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to extract cluster ID: %v", err)

		meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonClusterIDError,
			Message: fmt.Sprintf("Unable to extract cluster ID: %v", err),
		})
		if updateErr := r.Status().Update(ctx, cfg); updateErr != nil {
			logger.Error(updateErr, "failed to update status")
		}

		return ctrl.Result{}, fmt.Errorf("unable to extract the cluster ID from the namespace %q: %w", req.Namespace, err)
	}

	// Fabric Firewall Configuration Management
	fabricFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      forge.ForgeFabricResourceName(clusterID),
			Namespace: req.Namespace,
		},
	}

	fabricOp, err := controllerutil.CreateOrUpdate(ctx, r.Client, &fabricFwcfg, func() error {
		fabricFwcfg.SetLabels(forge.ForgeFabricLabels(clusterID))

		spec, err := forge.ForgeFabricSpec(ctx, r.Client, cfg, clusterID)
		if err != nil {
			return err
		}
		fabricFwcfg.Spec = *spec

		return controllerutil.SetOwnerReference(cfg, &fabricFwcfg, r.Scheme)
	})
	if err != nil {
		logger.Error(err, "unable to reconcile the fabric firewall configuration")

		r.Recorder.Eventf(cfg, corev1.EventTypeWarning, EventReasonReconcileError, "Failed to reconcile fabric: %v", err)

		meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeReady,
			Status:  metav1.ConditionFalse,
			Reason:  ReasonFabricSyncFailed,
			Message: fmt.Sprintf("Failed to sync FirewallConfiguration: %v", err),
		})
		if updateErr := r.Status().Update(ctx, cfg); updateErr != nil {
			logger.Error(updateErr, "failed to update status during error handling")
		}

		return ctrl.Result{}, fmt.Errorf("unable to reconcile the fabric firewall configuration: %w", err)
	}

	logger.Info("reconciliation completed", "fabricOp", fabricOp)

	// Success and Final Status Update
	cfg.Status.ObservedGeneration = cfg.Generation

	meta.SetStatusCondition(&cfg.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeReady,
		Status:  metav1.ConditionTrue,
		Reason:  ReasonFabricSynced,
		Message: "FirewallConfiguration successfully synced",
	})

	if err := r.Status().Update(ctx, cfg); err != nil {
		logger.Error(err, "failed to update PeeringSecurity status")
		return ctrl.Result{}, err
	}

	if fabricOp != controllerutil.OperationResultNone {
		r.Recorder.Eventf(cfg, corev1.EventTypeNormal, EventReasonSynced, "FirewallConfiguration %s successfully", fabricOp)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PeeringSecurityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO: watch network changes
	// TODO: watch pod changes
	// TODO: firewall configuration ownership

	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.PeeringSecurity{}).
		Named("peeringsecurity").
		Complete(r)
}
