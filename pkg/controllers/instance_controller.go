/*
Copyright 2021.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	equinixv1alpha1 "github.com/hobbyfarm/metal-operator/pkg/api/v1alpha1"
	"github.com/hobbyfarm/metal-operator/pkg/metal"
	"k8s.io/apimachinery/pkg/api/errors"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Threads int
	Log     logr.Logger
}

const instanceFinalizer = "instance.cattle.io"

//+kubebuilder:rbac:groups=equinix.cattle.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=equinix.cattle.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=equinix.cattle.io,resources=instances/finalizers,verbs=update

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	instance := &equinixv1alpha1.Instance{}

	var requeue bool
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// mClient contains the new metal client
	mClient, err := metal.NewClient(ctx, r.Client, instance.Spec.Secret, instance.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// instance provisioning //
		status := instance.Status.DeepCopy()

		switch status.Status {
		case "":
			// need to provision
			status, err = mClient.CreateNewDevice(instance)
		case "queued":
			// need to check if device is active
			status, err = mClient.CheckDeviceStatus(instance)
		case "active":
			// provisioning complete, update status and ignore
			return ctrl.Result{}, nil
		}

		if err != nil {
			return ctrl.Result{}, err
		}
		instance.Status = *status
		requeue = true
		controllerutil.AddFinalizer(instance, instanceFinalizer)
	} else {
		// handle termination of hardware //
		err = mClient.DeleteDevice(instance)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(instance, instanceFinalizer)
	}
	return ctrl.Result{Requeue: requeue}, r.Update(ctx, instance)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Threads,
		}).
		For(&equinixv1alpha1.Instance{}).
		Complete(r)
}
