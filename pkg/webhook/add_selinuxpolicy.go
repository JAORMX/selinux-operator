package webhook

import (
	"github.com/JAORMX/selinux-operator/pkg/webhook/namespace"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, namespace.Add)
}
