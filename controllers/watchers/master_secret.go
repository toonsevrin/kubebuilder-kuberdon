package watchers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/kuberty/kuberdon/api/v1beta1"
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


// MASTER SECRET (The root secret that will be deployed)
// Tested this manually and it works

// Add watcher that watches new namespaces (and deploy to them!)
type MasterSecretWatcher struct {
	Client client.Client
	Log    logr.Logger
}

// The master secret informer listens to all Registry (type: kubernetes.io/dockerconfigjson) secrets over all namespaces.
// It excludes registries with the label "kuberdon.kuberty.io: true"
func (s MasterSecretWatcher) Informer(kubernetesClientSet *kubernetes.Clientset) (informers.SharedInformerFactory, cache.Informer) {
	fieldSelector := fields.Set(map[string]string{"type": DockerconfigSecretType}).String()
	labelSelector := OwnedByKuberdonLabel + " != " + OwnedByKuberdonValue
	tweakListOptions := func(opts *metav1.ListOptions) {
		opts.FieldSelector = fieldSelector
		opts.LabelSelector = labelSelector
	}
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubernetesClientSet, time.Minute*30, informers.WithTweakListOptions(tweakListOptions))
	return sharedInformerFactory, sharedInformerFactory.Core().V1().Secrets().Informer()
}
func (s MasterSecretWatcher) ToRequestsFunc(mapObject handler.MapObject) []reconcile.Request {
	ctx := context.Background()
	// Format our secret name as to how it appears in the registry spec
	secretNamespacedName := types.NamespacedName{Name: mapObject.Meta.GetName(), Namespace: mapObject.Meta.GetNamespace()}.String()
	if mapObject.Meta.GetNamespace() == "" || mapObject.Meta.GetNamespace() == "default" {
		secretNamespacedName = mapObject.Meta.GetName()
	}
	// List all registries with the secret in their spec
	var registries v1beta1.RegistryList

	if err := s.Client.List(ctx, &registries, client.MatchingFields{SecretField: secretNamespacedName}); err != nil {
		s.Log.Error(err, "unable to list registries")
		return []reconcile.Request{}
	}

	// Return all the namespacednames
	requests := []reconcile.Request{} // We can improve performance by correctly setting the size of this array from the start
	for _, registry := range registries.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{Name: registry.Name, Namespace: registry.Namespace}})
	}

	return requests
}
func (s MasterSecretWatcher) GetIndices(mgr ctrl.Manager) []Index {
	return []Index{
		Index{
			Obj:   &v1beta1.Registry{},
			Field: SecretField,
			ExtractValue: func(obj runtime.Object) []string {
				registry := obj.(*v1beta1.Registry)
				return []string{registry.Spec.Secret}
			},
		},
	}
}
