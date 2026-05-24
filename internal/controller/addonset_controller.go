/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
*/

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	platformv1alpha1 "github.com/example/byo-addons/api/v1alpha1"
)

const (
	addonFinalizer = "addonset.platform.byoaddons.io/finalizer"
	managedLabel   = "platform.byoaddons.io/managed-by"
	componentLabel = "platform.byoaddons.io/component"
)

// AddonSetReconciler reconciles an AddonSet object.
type AddonSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.byoaddons.io,resources=addonsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.byoaddons.io,resources=addonsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.byoaddons.io,resources=addonsets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *AddonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var addonSet platformv1alpha1.AddonSet
	if err := r.Get(ctx, req.NamespacedName, &addonSet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if addonSet.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.AddFinalizer(&addonSet, addonFinalizer) {
			if err := r.Update(ctx, &addonSet); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(&addonSet, addonFinalizer) {
			if err := r.deleteDesiredStateConfigMaps(ctx, &addonSet); err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(&addonSet, addonFinalizer)
			return ctrl.Result{}, r.Update(ctx, &addonSet)
		}
		return ctrl.Result{}, nil
	}

	statuses := make([]platformv1alpha1.AddonComponentStatus, 0, len(addonSet.Spec.Components))
	ready := int32(0)
	desired := int32(0)

	for _, component := range addonSet.Spec.Components {
		if !component.Enabled {
			statuses = append(statuses, componentStatus(component, false, "disabled", ""))
			continue
		}

		desired++
		observed, err := r.reconcileComponentConfigMap(ctx, &addonSet, component)
		if err != nil {
			logger.Error(err, "failed to reconcile component", "component", component.Name)
			statuses = append(statuses, componentStatus(component, false, err.Error(), ""))
			continue
		}

		ready++
		statuses = append(statuses, componentStatus(component, true, "desired state recorded", observed))
	}

	addonSet.Status.ObservedGeneration = addonSet.Generation
	addonSet.Status.DesiredComponents = desired
	addonSet.Status.ReadyComponents = ready
	addonSet.Status.Components = statuses

	conditionStatus := metav1.ConditionFalse
	reason := "ComponentsPending"
	message := fmt.Sprintf("%d of %d enabled components reconciled", ready, desired)
	if desired == ready {
		conditionStatus = metav1.ConditionTrue
		reason = "ComponentsReconciled"
	}

	addonSet.Status.Conditions = upsertCondition(addonSet.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             conditionStatus,
		ObservedGeneration: addonSet.Generation,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})

	if err := r.Status().Update(ctx, &addonSet); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AddonSetReconciler) reconcileComponentConfigMap(ctx context.Context, addonSet *platformv1alpha1.AddonSet, component platformv1alpha1.AddonComponent) (string, error) {
	namespace := component.Namespace
	if namespace == "" {
		namespace = addonSet.Namespace
	}

	payload, err := json.MarshalIndent(component, "", "  ")
	if err != nil {
		return "", err
	}

	name := desiredStateName(addonSet.Name, component.Name)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		configMap.Labels = map[string]string{
			managedLabel:   "byo-addons-operator",
			componentLabel: component.Name,
		}
		configMap.Data = map[string]string{
			"addon.json": string(payload),
			"engine":     addonSet.Spec.GitOpsEngine,
			"cluster":    addonSet.Spec.ClusterName,
		}
		return controllerutil.SetControllerReference(addonSet, configMap, r.Scheme)
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("ConfigMap/%s in namespace %s", name, namespace), nil
}

func (r *AddonSetReconciler) deleteDesiredStateConfigMaps(ctx context.Context, addonSet *platformv1alpha1.AddonSet) error {
	for _, component := range addonSet.Spec.Components {
		namespace := component.Namespace
		if namespace == "" {
			namespace = addonSet.Namespace
		}

		configMap := &corev1.ConfigMap{}
		key := types.NamespacedName{Name: desiredStateName(addonSet.Name, component.Name), Namespace: namespace}
		if err := r.Get(ctx, key, configMap); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		if err := r.Delete(ctx, configMap); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func desiredStateName(addonSetName string, componentName string) string {
	name := strings.ToLower(fmt.Sprintf("%s-%s-desired", addonSetName, componentName))
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

func componentStatus(component platformv1alpha1.AddonComponent, ready bool, message string, observed string) platformv1alpha1.AddonComponentStatus {
	return platformv1alpha1.AddonComponentStatus{
		Name:      component.Name,
		Type:      component.Type,
		Provider:  component.Provider,
		Ready:     ready,
		Message:   message,
		Observed:  observed,
		UpdatedAt: metav1.Now(),
	}
}

func upsertCondition(conditions []metav1.Condition, next metav1.Condition) []metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == next.Type {
			if conditions[i].Status == next.Status {
				next.LastTransitionTime = conditions[i].LastTransitionTime
			}
			conditions[i] = next
			return conditions
		}
	}
	return append(conditions, next)
}

func (r *AddonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.AddonSet{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
