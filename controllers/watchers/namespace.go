package watchers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kuberty/kuberdon/api/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

//TODO: When a namespace is created or removed, the serviceAccount changes as well, we will always reconcile twice due to this.
// This watcher could thus be deleted if desired, all though this would contradict our "watch all resources relevant to the state" principle.
type NamespaceWatcher struct {
	Client client.Client
	Log    logr.Logger
}

// The master secret informer listens to all Registry (type: kubernetes.io/dockerconfigjson) secrets over all namespaces.
// It excludes registries with the label "kuberdon.kuberty.io: true"
func (s NamespaceWatcher) Informer(kubernetesClientSet *kubernetes.Clientset) (informers.SharedInformerFactory, cache.Informer) {
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubernetesClientSet, time.Minute*30)
	return sharedInformerFactory, sharedInformerFactory.Core().V1().Namespaces().Informer()
}

func (s NamespaceWatcher) ToRequestsFunc(mapObject handler.MapObject) []reconcile.Request {
	ctx := context.Background()

	// List all registries
	var registries v1beta1.RegistryList
	if err := s.Client.List(ctx, &registries); err != nil {
		s.Log.Error(err, "unable to list registries")
		return []reconcile.Request{}
	}

	// Return all the namespacednames
	var requests []reconcile.Request // We can improve performance by correctly setting the size of this array from the start
	for _, registry := range registries.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: registry.Name, Namespace: registry.Namespace}})
	}
	return requests
}

func (s NamespaceWatcher) GetIndices(mgr ctrl.Manager) []Index {
	return []Index{}
}
