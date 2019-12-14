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
the operator is running (`selinux-operator` by default).

TODO
====

* Create a validating webhook so that folks can only create pods with certain
  SELinux policies on namespaces that allow so.
