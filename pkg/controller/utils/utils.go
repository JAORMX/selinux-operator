package utils

import (
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
)

func GetPolicyName(name, ns string) string {
	return name + "-" + ns
}

// Remove "." from node names, which are invalid for pod names
func parseNodeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}

func GetInstallerPodName(name, ns string, node *corev1.Node) string {
	// policy-installer
	namePrefix := "p-i"
	parsedNodeName := parseNodeName(node.Name)
	podname := namePrefix + "-" + GetPolicyName(name, ns) + "-" + parsedNodeName

	// K8s has a 63 char name limit for pods
	if len(podname) > 62 {
		return podname[:62]
	}
	return podname
}

func GetPolicyConfigMapName(name, ns string) string {
	namePrefix := "policy-for"
	return namePrefix + "-" + GetPolicyName(name, ns)
}

// GetOperatorNamespace gets the namespace that the operator is currently running on.
func GetOperatorNamespace() string {
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return "openshift-selinux-operator"
	}
	return operatorNs
}
