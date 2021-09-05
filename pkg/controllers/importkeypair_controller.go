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
	"github.com/go-logr/logr"
	"github.com/hobbyfarm/metal-operator/pkg/metal"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	equinixv1alpha1 "github.com/hobbyfarm/metal-operator/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ImportKeyPairReconciler reconciles a ImportKeyPair object
type ImportKeyPairReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Threads int
	Log     logr.Logger
}

//+kubebuilder:rbac:groups=equinix.cattle.io,resources=importkeypairs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=equinix.cattle.io,resources=importkeypairs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=equinix.cattle.io,resources=importkeypairs/finalizers,verbs=update

func (r *ImportKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)

	// your logic here
	importKeyPair := &equinixv1alpha1.ImportKeyPair{}

	var requeue bool
	if err := r.Get(ctx, req.NamespacedName, importKeyPair); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// mClient contains the new metal client
	mClient, err := metal.NewClient(ctx, r.Client, importKeyPair.Spec.Secret, importKeyPair.Namespace)
	if err != nil {
		return ctrl.Result{}, err
	}

	if importKeyPair.ObjectMeta.DeletionTimestamp.IsZero() {
		status := importKeyPair.Status.DeepCopy()

		switch status.Status {
		case "":
			// create keypair
			status, err = mClient.CreateImportKeyPair(importKeyPair)
		case "created":
			return ctrl.Result{}, nil
		}

		if err != nil {
			return ctrl.Result{}, err
		}

		importKeyPair.Status = *status
		// always requeue since the processing ends in the switch block
		requeue = true
		controllerutil.AddFinalizer(importKeyPair, instanceFinalizer)
	} else {
		// handle termination of importKeyPair
		err = mClient.DeleteKeyPair(importKeyPair)
		if err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(importKeyPair, instanceFinalizer)
	}

	return ctrl.Result{Requeue: requeue}, r.Update(ctx, importKeyPair)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ImportKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: r.Threads,
		}).
		For(&equinixv1alpha1.ImportKeyPair{}).
		Complete(r)
}
