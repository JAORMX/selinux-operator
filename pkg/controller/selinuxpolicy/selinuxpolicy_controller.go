package selinuxpolicy

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
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

// The underscore is not a valid character in a pod, so we can
// safely use it as a separator.
const policyWrapper = `(block {{.Name}}_{{.Namespace}}
    {{.Policy}}
)`

const selinuxFinalizerName = "selinuxpolicy.finalizers.selinuxpolicy.openshift.io"

// Add creates a new SelinuxPolicy Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// Create template to wrap policies
	tmpl, _ := template.New("policyWrapper").Parse(policyWrapper)
	return &ReconcileSelinuxPolicy{client: mgr.GetClient(), scheme: mgr.GetScheme(), policyTemplate: tmpl}
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
	client         client.Client
	scheme         *runtime.Scheme
	policyTemplate *template.Template
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
		return reconcile.Result{}, utils.IgnoreNotFound(err)
	}

	policyCopy := instance.DeepCopy()
	if policyCopy.Status.State == "" {
		policyCopy.Status.State = selinuxv1alpha1.PolicyStatePending
		if err := r.client.Status().Update(context.TODO(), policyCopy); err != nil {
			return reconcile.Result{}, err
		}
	}

	// If "apply" is false, no need to do anything, let the deployer
	// review it.
	if !instance.Spec.Apply {
		policyCopy := instance.DeepCopy()
		policyCopy.Status.State = selinuxv1alpha1.PolicyStatePending
		if err := r.client.Status().Update(context.TODO(), policyCopy); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if instance.Status.Usage != utils.GetPolicyUsage(instance.Name, instance.Namespace) {
		if err = r.addUsageStatus(instance, reqLogger); err != nil {
			return reconcile.Result{}, err
		}
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted
		if !utils.SliceContainsString(instance.ObjectMeta.Finalizers, selinuxFinalizerName) {
			return r.addFinalizer(instance, reqLogger)
		}
		return r.reconcileConfigMap(instance, reqLogger)
	} else {
		// The object is being deleted
		if utils.SliceContainsString(instance.ObjectMeta.Finalizers, selinuxFinalizerName) {
			if err := r.deleteConfigMap(instance, reqLogger); err != nil {
				return reconcile.Result{}, err
			}

			return r.removeFinalizer(instance, reqLogger)
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileSelinuxPolicy) addFinalizer(sp *selinuxv1alpha1.SelinuxPolicy, logger logr.Logger) (reconcile.Result, error) {
	spcopy := sp.DeepCopy()
	spcopy.ObjectMeta.Finalizers = append(spcopy.ObjectMeta.Finalizers, selinuxFinalizerName)
	if err := r.client.Update(context.Background(), spcopy); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSelinuxPolicy) addUsageStatus(sp *selinuxv1alpha1.SelinuxPolicy, logger logr.Logger) error {
	spcopy := sp.DeepCopy()
	spcopy.Status.Usage = utils.GetPolicyUsage(spcopy.Name, spcopy.Namespace)
	if err := r.client.Status().Update(context.Background(), spcopy); err != nil {
		return err
	}
	return nil
}

func (r *ReconcileSelinuxPolicy) removeFinalizer(sp *selinuxv1alpha1.SelinuxPolicy, logger logr.Logger) (reconcile.Result, error) {
	spcopy := sp.DeepCopy()
	spcopy.ObjectMeta.Finalizers = utils.RemoveStringFromSlice(spcopy.ObjectMeta.Finalizers, selinuxFinalizerName)
	if err := r.client.Update(context.Background(), spcopy); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSelinuxPolicy) reconcileConfigMap(instance *selinuxv1alpha1.SelinuxPolicy, logger logr.Logger) (reconcile.Result, error) {
	// Define a new ConfigMap object
	cm := r.newConfigMapForPolicy(instance)

	// Check if this cm already exists
	foundCM := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, foundCM)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
		if err = r.client.Create(context.TODO(), cm); err != nil {
			return reconcile.Result{}, utils.IgnoreAlreadyExists(err)
		}

		// CM created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileSelinuxPolicy) deleteConfigMap(instance *selinuxv1alpha1.SelinuxPolicy, logger logr.Logger) error {
	// Define a new ConfigMap object
	cm := r.newConfigMapForPolicy(instance)
	logger.Info("Deleting ConfigMap", "ConfigMap.Namespace", cm.Namespace, "ConfigMap.Name", cm.Name)
	return utils.IgnoreNotFound(r.client.Delete(context.TODO(), cm))
}

func (r *ReconcileSelinuxPolicy) newConfigMapForPolicy(cr *selinuxv1alpha1.SelinuxPolicy) *corev1.ConfigMap {
	labels := map[string]string{
		"appName":      cr.Name,
		"appNamespace": cr.Namespace,
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.GetPolicyConfigMapName(cr.Name, cr.Namespace),
			Namespace: utils.GetOperatorNamespace(),
			Labels:    labels,
		},
		Data: map[string]string{
			utils.GetPolicyName(cr.Name, cr.Namespace) + ".cil": r.wrapPolicy(cr),
		},
	}
}

func (r *ReconcileSelinuxPolicy) wrapPolicy(cr *selinuxv1alpha1.SelinuxPolicy) string {
	parsedpolicy := strings.TrimSpace(cr.Spec.Policy)
	// ident
	parsedpolicy = strings.ReplaceAll(parsedpolicy, "\n", "\n    ")
	// replace empty lines
	parsedpolicy = strings.TrimSpace(parsedpolicy)
	data := struct {
		Name      string
		Namespace string
		Policy    string
	}{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Policy:    parsedpolicy,
	}
	var result bytes.Buffer
	r.policyTemplate.Execute(&result, data)
	return result.String()
}
