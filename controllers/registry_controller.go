/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	kuberdonv1beta1 "github.com/kuberty/kuberdon/api/v1beta1"
	"github.com/kuberty/kuberdon/controllers/watchers"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"strings"
	"time"
)

const (
	childSecretPrefix = "kuberty-%s"
	childSecretNameFormat     = childSecretPrefix + "-%s"
	nameField                 = "metadata.name"
	defaultServiceAccountName = "default"
)

const (
	SyncedState = "Synced"
	SyncedMessage = "All secrets deployed"
	ErrorState = "Error"
	ErrorMessage = "Error %v: retrying..."
)

var (
	faultyNamespaceRegexError   = "BadNamespaceFilterError"
	faultyNamespaceRegexMessage = "%v namespaceFilters are not regex."
	resolver                    = DefaultNamespaceResolver{}
)

// KuberdonReconciler reconciles a Kuberdon object
type KuberdonReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *KuberdonReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("testresource", req.NamespacedName)

	reconciliation := &Reconciliation{Client: r.Client, Context: ctx}

	// Fetch the registry
	if err := r.Client.Get(ctx, req.NamespacedName, &reconciliation.Registry); err != nil {
		log.Error(err, "Unable to fetch Registry")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//fetch all namespaces
	if err := r.Client.List(ctx, &reconciliation.Namespaces); err != nil {
		log.Error(err, "Unable to fetch namespaces")
		return ctrl.Result{}, err
	}

	//fetch the actual and desired state
	states, err := reconciliation.getActualAndDesiredState(resolver)
	if err != nil {
		log.Error(err, "Unable to fetch states")
		return ctrl.Result{}, err
	}
	//reconcile current and desired
	err = r.executeReconcile(ctx, states.Actual, states.Desired)
	if err != nil {
		log.Error(err, "Unable to execute reconciliation")
		newReg := reconciliation.Registry.DeepCopy()
		newReg.Status = kuberdonv1beta1.RegistryStatus{State: kuberdonv1beta1.ErrorState, Message:fmt.Sprintf(ErrorMessage, err)}
		if err2 := r.Status().Update(ctx, newReg); err2 != nil {
			log.Error(err2, "Unable to update status")
			return ctrl.Result{Requeue:true}, err2
		}
		return ctrl.Result{Requeue:true,RequeueAfter: time.Second * 10}, err
	}
	if reconciliation.Registry.Status.State != kuberdonv1beta1.SyncedState {
		newReg := reconciliation.Registry.DeepCopy()
		newReg.Status = kuberdonv1beta1.RegistryStatus{State: kuberdonv1beta1.SyncedState, Message:SyncedMessage}
		if err := r.Status().Update(ctx, newReg); err != nil {
			log.Error(err, "Unable to update status")
			return ctrl.Result{Requeue:true}, err
		}
	}
	return ctrl.Result{}, nil
}

func namespacedName(namespacedString string) types.NamespacedName {
	namespacedName := types.NamespacedName{Namespace: "default", Name: namespacedString}
	if strings.Contains(namespacedString, "/") {
		sections := strings.Split(namespacedString, "/")
		namespacedName.Namespace = sections[0]
		namespacedName.Name = sections[1]
	}
	return namespacedName
}

func getChildSecretName(controllerName string, secretName string) string {
	return fmt.Sprintf(childSecretNameFormat, controllerName, secretName)
}

// +kubebuilder:rbac:groups=kuberdon.kuberty.io,resources=kuberdons,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kuberdon.kuberty.io,resources=kuberdons/status,verbs=get;update;patch

func (r *KuberdonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	generatedClient := kubernetes.NewForConfigOrDie(mgr.GetConfig())

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&kuberdonv1beta1.Registry{})

	// Watchers watch the kubernetes API for changes, they are defined in watchers/watchers.go.
	// Whenever the master secret (the one that gets deployed) or the child secrets (the ones that are deployed) change,
	// the reconcile function will be called for the relevant Registries.
	watchersProvider := watchers.AllWatchers{Client: r.Client, Log: r.Log}
	watchers := watchersProvider.GetWatchers()

	// Let's add the watchers to the manager and the builder.
	for _, watcher := range watchers {
		informer, err := addWatcherToMgr(mgr, generatedClient, watcher)
		if err != nil {
			return err
		}
		builder = builder.Watches(&source.Informer{Informer: informer}, &handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(watcher.ToRequestsFunc)})
	}

	return builder.Complete(r)
}


func addWatcherToMgr(mgr ctrl.Manager, clientSet *kubernetes.Clientset, watcher watchers.Watcher) (cache.Informer, error) {
	informerFactory, informer := watcher.Informer(clientSet)

	for _, index := range watcher.GetIndices(mgr) {
		mgr.GetFieldIndexer().IndexField(index.Obj, index.Field, index.ExtractValue)
	}

	return informer, mgr.Add(manager.RunnableFunc(func(s <-chan struct{}) error {
		informerFactory.Start(s)
		<-s
		return nil
	}))
}
