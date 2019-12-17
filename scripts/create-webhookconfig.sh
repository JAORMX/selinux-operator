#!/bin/bash

oc create configmap temp
oc annotate configmap temp service.beta.openshift.io/inject-cabundle="true"
oc get configmap/temp -o jsonpath='{.data.service-ca\.crt}' > /tmp/ca-bundle.pem

cat << EOF | oc create -f -
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: "selinux-policy-in-pod-namespace.openshift.io"
webhooks:
- name: "selinux-policy-in-pod-namespace.openshift.io"
  rules:
  - apiGroups:   [""]
    apiVersions: ["v1"]
    operations:  ["CREATE", "UPDATE"]
    resources:   ["pods"]
    scope:       "Namespaced"
  clientConfig:
    service:
      namespace: "openshift-selinux-operator"
      name: "selinux-namespace-webhook"
      path: "/validate-selinuxpolicy-namespace"
      port: 8443
    caBundle: "$(base64 -w0 /tmp/ca-bundle.pem)"
  admissionReviewVersions: ["v1beta1"]
  sideEffects: None
  timeoutSeconds: 2
EOF

oc delete configmap temp
