/*
Copyright 2019 Red Hat Inc.

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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var selinuxpolicylog = logf.Log.WithName("selinuxpolicy-resource")

func AddWebhook(mgr manager.Manager) error {
	if err := (&SelinuxPolicy{}).SetupWebhookWithManager(mgr); err != nil {
		selinuxpolicylog.Error(err, "unable to create webhook", "webhook", "SelinuxPolicy")
		return err
	}
	return nil
}

func (r *SelinuxPolicy) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-selinux-openshift-io-v1alpha1-selinuxpolicy,mutating=false,failurePolicy=fail,groups=selinux.openshift.io,resources=selinuxpolicies,versions=v1alpha1,name=vselinuxpolicy.kb.io

var _ webhook.Validator = &SelinuxPolicy{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *SelinuxPolicy) ValidateCreate() error {
	selinuxpolicylog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *SelinuxPolicy) ValidateUpdate(old runtime.Object) error {
	selinuxpolicylog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *SelinuxPolicy) ValidateDelete() error {
	selinuxpolicylog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
