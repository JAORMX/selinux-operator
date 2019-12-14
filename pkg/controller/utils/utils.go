package utils

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"

	selinuxv1alpha1 "github.com/JAORMX/selinux-operator/pkg/apis/selinux/v1alpha1"
)

func GetPolicyName(cr *selinuxv1alpha1.SelinuxPolicy) string {
	return cr.Name + "-" + cr.Namespace
}

func GetInstallerPodName(cr *selinuxv1alpha1.SelinuxPolicy, node *corev1.Node) string {
	namePrefix := "policy-installer"
	return namePrefix + "-" + GetPolicyName(cr) + "-" + node.Name
}

func GetPolicyConfigMapName(cr *selinuxv1alpha1.SelinuxPolicy) string {
	namePrefix := "policy-for"
	return namePrefix + "-" + GetPolicyName(cr)
}

// GetOperatorNamespace gets the namespace that the operator is currently running on.
func GetOperatorNamespace() string {
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return "selinux-operator"
	}
	return operatorNs
}
