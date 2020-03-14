package watchers

import (
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
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

// CHILD SECRET (The owned secrets): This could also be done with the Owns method of the builder, we'll leave this here as it allows a bit more optimization (all though there is more boilerplate)

type ChildSecretWatcher struct {
	Client client.Client
	Log    logr.Logger
}

func (s ChildSecretWatcher) Informer(kubernetesClientSet *kubernetes.Clientset) (informers.SharedInformerFactory, cache.Informer) {
	fieldSelector := fields.Set(map[string]string{"type": DockerconfigSecretType}).String()
	labelSelector := labels.Set(map[string]string{OwnedByKuberdonLabel: OwnedByKuberdonValue}).String()
	tweakListOptions := func(opts *metav1.ListOptions) {
		opts.FieldSelector = fieldSelector
		opts.LabelSelector = labelSelector
	}
	sharedInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubernetesClientSet, time.Minute*30, informers.WithTweakListOptions(tweakListOptions))
	return sharedInformerFactory, sharedInformerFactory.Core().V1().Secrets().Informer()
}

func (s ChildSecretWatcher) ToRequestsFunc(mapObject handler.MapObject) []reconcile.Request {
	for _, ownerReference := range mapObject.Meta.GetOwnerReferences() {
		if (*ownerReference.Controller) == true {
			if ownerReference.APIVersion != apiGVStr || ownerReference.Kind != "registryKind" {
				return []reconcile.Request{}
			}
			return []reconcile.Request{reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ownerReference.Name},
			},
			}
		}
	}
	return []reconcile.Request{}
}

func (s ChildSecretWatcher) GetIndices(mgr ctrl.Manager) []Index {
	return []Index{
		Index{
			Obj:   &v1.Secret{},
			Field: OwningControllerField,
			ExtractValue: func(obj runtime.Object) []string {
				secret := obj.(*v1.Secret)

				owner := metav1.GetControllerOf(secret)
				if owner == nil {
					return nil
				}
				// ...make sure the secret is owned by a Register...
				if owner.APIVersion != apiGVStr || owner.Kind != RegistryKind {
					return nil
				}

				// ...and if so, return it
				return []string{owner.Name}
			},
		},
	}
}
