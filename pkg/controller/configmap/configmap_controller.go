package configmap

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	selinuxv1alpha1 "github.com/JAORMX/selinux-operator/pkg/apis/selinux/v1alpha1"
	"github.com/JAORMX/selinux-operator/pkg/controller/utils"
)

var log = logf.Log.WithName("controller_configmap")

// Add creates a new ConfigMap Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConfigMap{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("configmap-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ConfigMap
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ConfigMap
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.ConfigMap{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileConfigMap implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConfigMap{}

// ReconcileConfigMap reconciles a ConfigMap object
type ReconcileConfigMap struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ConfigMap object and makes changes based on the state read
// and what is in the ConfigMap.Spec
func (r *ReconcileConfigMap) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Skip ConfigMaps from other namespaces
	if request.Namespace != utils.GetOperatorNamespace() {
		return reconcile.Result{}, nil
	}

	// Fetch the ConfigMap instance
	cminstance := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cminstance)
	if err != nil {
		return reconcile.Result{}, utils.IgnoreNotFound(err)
	}
	policyName, ok := cminstance.Labels["appName"]
	if !ok {
		return reconcile.Result{}, nil
	}
	policyNamespace, ok := cminstance.Labels["appNamespace"]
	if !ok {
		return reconcile.Result{}, nil
	}

	reqLogger := log.WithValues("SelinuxPolicy.Name", policyName, "SelinuxPolicy.Namespace", policyNamespace)

	policyObjKey := types.NamespacedName{Name: policyName, Namespace: policyNamespace}
	policy := &selinuxv1alpha1.SelinuxPolicy{}
	err = r.client.Get(context.TODO(), policyObjKey, policy)
	if err != nil {
		return reconcile.Result{}, utils.IgnoreNotFound(err)
	}
	policyCopy := policy.DeepCopy()

	if policyCopy.Status.State == "" || policyCopy.Status.State == selinuxv1alpha1.PolicyStatePending {
		policyCopy.Status.State = selinuxv1alpha1.PolicyStateInProgress
		r.client.Status().Update(context.TODO(), policyCopy)
		// Create another copy so we don't modify the cache
		policyCopy = policy.DeepCopy()
	}
	reqLogger.Info("Reconciling pods for policy")

	nodesList := &corev1.NodeList{}
	foundPods := []*corev1.Pod{}
	err = r.client.List(context.TODO(), nodesList)
	for _, node := range nodesList.Items {
		// Define a new Pod object
		pod := newPodForPolicy(policyName, policyNamespace, &node)
		if err = controllerutil.SetControllerReference(cminstance, pod, r.scheme); err != nil {
			log.Error(err, "Failed to set pod ownership", "pod", pod)
			return reconcile.Result{}, err
		}

		// Check if this Pod already exists
		found := &corev1.Pod{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			if err = r.client.Create(context.TODO(), pod); err != nil {
				return reconcile.Result{}, utils.IgnoreAlreadyExists(err)
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// Pod already exists - don't requeue
		foundPods = append(foundPods, found)
	}

	// Lets check the state of the pods now
	for _, pod := range foundPods {
		exitCode, found := r.getInstallerContainerExitCode(pod)
		// Pod is still running - requeue
		if !found {
			reqLogger.Info("Pod still running", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return reconcile.Result{Requeue: true, RequeueAfter: 5 * time.Second}, nil
		}
		if exitCode != 0 {
			reqLogger.Info("Pod failed", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name, "exit-code", exitCode)
			policyCopy.Status.State = selinuxv1alpha1.PolicyStateError
			if err := r.client.Status().Update(context.TODO(), policyCopy); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	}
	policyCopy.Status.State = selinuxv1alpha1.PolicyStateInstalled
	if err := r.client.Status().Update(context.TODO(), policyCopy); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfigMap) getInstallerContainerExitCode(pod *corev1.Pod) (int32, bool) {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == "policy-installer" {
			if containerStatus.State.Terminated != nil {
				termState := containerStatus.State.Terminated
				return termState.ExitCode, true
			}
		}
	}
	return -1, false
}

// newPodForPolicy returns a busybox pod with the same name/namespace as the cr
func newPodForPolicy(name, ns string, node *corev1.Node) *corev1.Pod {
	//namespace := "selinux-policy-helper-operator"
	labels := map[string]string{
		"appName":      name,
		"appNamespace": ns,
	}
	trueVal := true
	hostVolTypeDir := corev1.HostPathDirectory
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetInstallerPodName(name, ns, node),
			Namespace: utils.GetOperatorNamespace(),
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "policy-installer",
					Image:   "quay.io/jaosorior/udica",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "semodule -vi /tmp/policy/*.cil /usr/share/udica/templates/*cil;"},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{"/bin/sh", "-c", fmt.Sprintf("semodule -vr '%s'", utils.GetPolicyName(name, ns))},
							},
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &trueVal,
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name:      "fsselinux",
							MountPath: "/sys/fs/selinux",
						},
						corev1.VolumeMount{
							Name:      "etcselinux",
							MountPath: "/etc/selinux",
						},
						corev1.VolumeMount{
							Name:      "varlibselinux",
							MountPath: "/var/lib/selinux",
						},
						corev1.VolumeMount{
							Name:      "policyvolume",
							MountPath: "/tmp/policy",
						},
					},
				},
				// This container needs to keep running so we can run the uninstall script.
				{
					Name:    "policy-uninstaller",
					Image:   "quay.io/jaosorior/udica",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", "while true; do sleep 30; done;"},
					Lifecycle: &corev1.Lifecycle{
						PreStop: &corev1.Handler{
							Exec: &corev1.ExecAction{
								Command: []string{"/bin/sh", "-c", fmt.Sprintf("semodule -vr '%s'", utils.GetPolicyName(name, ns))},
							},
						},
					},
					SecurityContext: &corev1.SecurityContext{
						Privileged: &trueVal,
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name:      "fsselinux",
							MountPath: "/sys/fs/selinux",
						},
						corev1.VolumeMount{
							Name:      "etcselinux",
							MountPath: "/etc/selinux",
						},
						corev1.VolumeMount{
							Name:      "varlibselinux",
							MountPath: "/var/lib/selinux",
						},
						corev1.VolumeMount{
							Name:      "policyvolume",
							MountPath: "/tmp/policy",
						},
					},
				},
			},
			ServiceAccountName: "selinux-operator",
			RestartPolicy:      corev1.RestartPolicyNever,
			NodeName:           node.Name,
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "fsselinux",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/sys/fs/selinux",
							Type: &hostVolTypeDir,
						},
					},
				},
				corev1.Volume{
					Name: "etcselinux",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/selinux",
							Type: &hostVolTypeDir,
						},
					},
				},
				corev1.Volume{
					Name: "varlibselinux",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/lib/selinux",
							Type: &hostVolTypeDir,
						},
					},
				},
				corev1.Volume{
					Name: "policyvolume",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: utils.GetPolicyConfigMapName(name, ns),
							},
						},
					},
				},
			},
			Tolerations: []corev1.Toleration{
				{
					Key:      "node-role.kubernetes.io/master",
					Operator: "Exists",
					Effect:   "NoSchedule",
				},
			},
		},
	}
}
