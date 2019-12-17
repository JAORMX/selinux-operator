package selinuxpolicy

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	selinuxv1alpha1 "github.com/JAORMX/selinux-operator/pkg/apis/selinux/v1alpha1"
	"github.com/JAORMX/selinux-operator/pkg/controller/utils"
)

var log = logf.Log.WithName("controller_selinuxpolicy")

// Add creates a new SelinuxPolicy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSelinuxPolicy{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("selinuxpolicy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SelinuxPolicy
	err = c.Watch(&source.Kind{Type: &selinuxv1alpha1.SelinuxPolicy{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSelinuxPolicy implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSelinuxPolicy{}

// ReconcileSelinuxPolicy reconciles a SelinuxPolicy object
type ReconcileSelinuxPolicy struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SelinuxPolicy object and makes changes based on the state read
// and what is in the SelinuxPolicy.Spec
func (r *ReconcileSelinuxPolicy) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling SelinuxPolicy")

	// Fetch the SelinuxPolicy instance
	instance := &selinuxv1alpha1.SelinuxPolicy{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("selinux policy instance not found", "Name", instance.Name, "Namespace", instance.Namespace)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new ConfigMap object
	cm := newConfigMapForPolicy(instance)

	// Check if this cm already exists
	foundCM := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return reconcile.Result{}, err
		}

		// CM created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("Reconcile: ConfigMap already exists", "ConfigMap.Namespace", foundCM.Namespace, "ConfigMap.Name", foundCM.Name)

	return reconcile.Result{}, nil
}

func newConfigMapForPolicy(cr *selinuxv1alpha1.SelinuxPolicy) *corev1.ConfigMap {
	labels := map[string]string{
		"appName":      cr.Name,
		"appNamespace": cr.Namespace,
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetPolicyConfigMapName(cr),
			Namespace: utils.GetOperatorNamespace(),
			Labels:    labels,
		},
		Data: map[string]string{
			utils.GetPolicyName(cr) + ".cil": cr.Spec.Policy,
		},
	}
}
