package apis

import (
	selinuxpolicy "github.com/JAORMX/selinux-operator/pkg/apis/selinux/v1alpha1"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddWebhooksToManagerFuncs = append(AddWebhooksToManagerFuncs, selinuxpolicy.AddWebhook)
}
