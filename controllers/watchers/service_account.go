package watchers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kuberty/kuberdon/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
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

type ServiceAccountWatcher struct {
	Client client.Client
	Log    logr.Logger
}

const (
	serviceAccountNameField   = "metadata.name"
	defaultServiceAccountName = "default"
)

// The master secret informer listens to all Registry (type: kubernetes.io/dockerconfigjson) secrets over all namespaces.
// It excludes registries with the label "kuberdon.kuberty.io: true"
func (s ServiceAccountWatcher) Informer(kubernetesClientSet *kubernetes.Clientset) (informers.SharedInformerFactory, cache.Informer) {
	fieldSelector := fields.Set(map[string]string{serviceAccountNameField: defaultServiceAccountName}).String()
	tweakListOptions := func(opts *metav1.ListOptions) {
		opts.FieldSelector = fieldSelector
	}
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubernetesClientSet, time.Minute*30, informers.WithTweakListOptions(tweakListOptions))
	return sharedInformerFactory, sharedInformerFactory.Core().V1().ServiceAccounts().Informer()
}

func (s ServiceAccountWatcher) ToRequestsFunc(mapObject handler.MapObject) []reconcile.Request {
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

func (s ServiceAccountWatcher) GetIndices(mgr ctrl.Manager) []Index {
	return []Index{
		Index{
			Obj:   &v1.ServiceAccount{},
			Field: serviceAccountNameField,
			ExtractValue: func(obj runtime.Object) []string {
				registry := obj.(*v1.ServiceAccount)
				return []string{registry.Name}
			},
		},
	}
}
