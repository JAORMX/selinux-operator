package utils

import (
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
)

// GetPolicyName gets the policy module name in the format that
// we're expecting for parsing.
func GetPolicyName(name, ns string) string {
	return name + "_" + ns
}

// GetPolicyK8sName gets the policy name in a format that's OK for k8s names.
func GetPolicyK8sName(name, ns string) string {
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
	podname := namePrefix + "-" + GetPolicyK8sName(name, ns) + "-" + parsedNodeName

	// K8s has a 63 char name limit for pods
	if len(podname) > 62 {
		return podname[:62]
	}
	return podname
}

func GetPolicyConfigMapName(name, ns string) string {
	namePrefix := "policy-for"
	return namePrefix + "-" + GetPolicyK8sName(name, ns)
}

// GetOperatorNamespace gets the namespace that the operator is currently running on.
func GetOperatorNamespace() string {
	operatorNs, err := k8sutil.GetOperatorNamespace()
	if err != nil {
		return "openshift-selinux-operator"
	}
	return operatorNs
}
