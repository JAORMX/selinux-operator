selinux-operator
================

This is a continuation to the work that I started on:

https://github.com/JAORMX/selinux-policy-helper-operator

It implements an operator that tracks for the creation of CustomResources
called `SelinuxPolicy`. These custom resources represent an SELinux policy that
can be installed in the system and that's available in a certain namespace.
So... they are namespaced resources.

The operator will listen for `SelinuxPolicy` objects on all namespaces of the
cluster, and if they exist, it'll create a `ConfigMap` on the namespace where
the operator is running (`selinux-operator` by default). And create a pod that
will install the policy in the node where it's running.

Once the `SelinuxPolicy` object is deleted, the policy removal will be
triggered via a finalizer that's installed on the pod.

A validating webhook is also implemented. It listens for all pods created in
the cluster, and will verify that the selinux policy (specified in the
`SELinuxOptions` section of the `securityContext`) corresponds to a
`SelinuxPolicy` object that exists in the namespace of the pod.

Installation instructions
-------------------------

Create CRD:

```
oc create -f deploy/crds/selinux.openshift.io_selinuxpolicies_crd.yaml
```

Deploy operator

```
oc create -f deploy/
```

Deploy webhook

```
./scripts/create-webhookconfig.sh
```
