package controllers

import (
	"context"
	"fmt"
	"github.com/kuberty/kuberdon/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

type State struct {
	MasterSecret *v1.Secret
	ChildSecrets []*v1.Secret
	Namespaces   []*v1.Namespace
}

type NamespaceState struct{
	Namespace *v1.Namespace
	ServiceAcccount *v1.ServiceAccount
}

func (r *KuberdonReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("testresource", req.NamespacedName)

	// Fetch the registry
	var registry v1beta1.Registry
	if err := r.Client.Get(ctx, req.NamespacedName, &registry); err != nil {
		log.Error(err, "Unable to fetch TestResource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//fetch the actual state

	//fetch the desired state

	//reconcile current and desired
	return ctrl.Result{}, nil
}

func getActualState(ctx context.Context, registry v1beta1.Registry, client client.Client) (State, error) {
	state := State{}
	//fetch the master secret
	if err := client.Get(ctx, namespacedName(registry.Spec.Secret), state.MasterSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return State{}, err
		}
	}
	//fetch the childSecrets

	return state, nil
}


func namespacedName(namespacedString string) types.NamespacedName {
	namespacedName := types.NamespacedName{Name: namespacedString}
	if strings.Contains(namespacedString, "/") {
		sections := strings.Split(namespacedString, "/")
		namespacedName.Namespace = sections[0]
		namespacedName.Name = sections[1]
	}
	return namespacedName
}

func getChildSecretName(controllerName string, secretName string) string{
	return fmt.Sprintf("kuberty-%s-%s", controllerName, secretName)
}
//func handleSecretNotFound(ctx context.Context, r *KuberdonReconciler,registry v1beta1.Registry) (ctrl.Result, error) {
//	ctx := context.Background()
//	log := r.Log.WithValues("testresource", req.NamespacedName)
//
//	// Fetch the registry
//	var registry v1beta1.Registry
//	if err := r.Client.Get(ctx, req.NamespacedName, &registry); err != nil {
//		log.Error(err, "Unable to fetch TestResource")
//		return ctrl.Result{}, client.IgnoreNotFound(err)
//	}
//	// Fetch the secret
//	var secret v1.Secret
//	if err := r.Client.Get(ctx, registry.Spec.Secret, secret); err != nil {
//		log.Error(err, "Unable to fetch Client Secret")
//		if apierrors.IsNotFound(err){
//
//		}else {
//			return ctrl.Result{}, err
//		}
//		return ctrl.Result{}, client.IgnoreNotFound(err)
//
//	}
//
//
//	return ctrl.Result{}, nil
//}
