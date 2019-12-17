package apis

import (
	admissionv1 "k8s.io/api/admission/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, admissionv1beta1.AddToScheme)
	AddToSchemes = append(AddToSchemes, admissionv1.AddToScheme)
}
