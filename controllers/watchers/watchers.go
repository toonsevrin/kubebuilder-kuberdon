package watchers

// TODO: Automated testing through [Kuberty](https://github.com/kuberty/kuberty)

import (
	"github.com/go-logr/logr"
	"github.com/kuberty/kuberdon/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//Todo: Split these up into a globals file or something
const (
	OwnedByKuberdonLabel   = "kuberdon.kuberty.io"
	OwnedByKuberdonValue   = "true"
	SecretField            = "spec.secret"
	DockerconfigSecretType = "kubernetes.io/dockerconfigjson"

	OwningControllerField = ".metadata.controller"

	RegistryKind = "Registry"
)

type Provider interface {
	GetWatchers() []Watcher
}
type AllWatchers struct {
	Client client.Client
	Log    logr.Logger
}

func (s AllWatchers) GetWatchers() []Watcher {
	return []Watcher{
		ChildSecretWatcher{Client: s.Client, Log: s.Log},
		MasterSecretWatcher{Client: s.Client, Log: s.Log},
		ServiceAccountWatcher{Client: s.Client, Log: s.Log},
		NamespaceWatcher{Client: s.Client, Log: s.Log},
	}
}

var apiGVStr = v1beta1.GroupVersion.String()

type Watcher interface {
	Informer(kubernetesClientSet *kubernetes.Clientset) (informers.SharedInformerFactory, cache.Informer)
	ToRequestsFunc(mapObject handler.MapObject) []reconcile.Request
	GetIndices(mgr ctrl.Manager) []Index
}
type Index struct {
	Obj          runtime.Object
	Field        string
	ExtractValue client.IndexerFunc
}
