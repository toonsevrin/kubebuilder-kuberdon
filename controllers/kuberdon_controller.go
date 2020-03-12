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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	kuberdonv1beta1 "github.com/kuberty/kuberdon/api/v1beta1"
	"github.com/kuberty/kuberdon/controllers/watchers"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KuberdonReconciler reconciles a Kuberdon object
type KuberdonReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
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
	watchers := []watchers.Watcher{
		watchers.MasterSecretWatcher{Client: r.Client, Log: r.Log},
		watchers.ChildSecretWatcher{},
	}

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
