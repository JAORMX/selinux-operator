package utils

import (
	hash "crypto/sha1"
	"fmt"
	"io"
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// GetPolicyName gets the policy module name in the format that
// we're expecting for parsing.
func GetPolicyName(name, ns string) string {
	return name + "_" + ns
}

// GetPolicyUsage is the representation of how a pod will call this
// SELinux module
func GetPolicyUsage(name, ns string) string {
	return GetPolicyName(name, ns) + ".process"
}

// GetPolicyK8sName gets the policy name in a format that's OK for k8s names.
func GetPolicyK8sName(name, ns string) string {
	return name + "-" + ns
}

// Remove "." from node names, which are invalid for pod names
func parseNodeName(name string) string {
	return strings.ReplaceAll(name, ".", "-")
}

// GetInstallerPodName gets the name of the installer pod. Given that the pod names
// can get pretty long, we hash the name so it fits in the space and is also
// unique.
func GetInstallerPodName(name, ns string, node *corev1.Node) string {
	// policy-installer
	parsedNodeName := parseNodeName(node.Name)
	podname := GetPolicyK8sName(name, ns) + "-" + parsedNodeName

	hasher := hash.New()
	io.WriteString(hasher, podname)
	return fmt.Sprintf("%x", hasher.Sum(nil))
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

// SliceContainsString helper function to check if a string is in a slice of strings
func SliceContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveStringFromSlice helper function to remove a string from a slice
func RemoveStringFromSlice(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// IgnoreNotFound ignores "NotFound" errors
func IgnoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}

// IgnoreAlreadyExists ignores "AlreadyExists" errors
func IgnoreAlreadyExists(err error) error {
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
