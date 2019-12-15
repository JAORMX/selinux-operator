package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}

// AddWebhooksToManagerFuncs registers all the webhook registration functions
var AddWebhooksToManagerFuncs []func(manager.Manager) error

// AddWebhooksToManager adds all Webhooks to the Manager
func AddWebhooksToManager(m manager.Manager) error {
	for _, f := range AddWebhooksToManagerFuncs {
		if err := f(m); err != nil {
			return err
		}
	}
	return nil
}
