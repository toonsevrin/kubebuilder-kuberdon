package controllers

import (
	"context"
	"fmt"
	"github.com/kuberty/kuberdon/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	childSecretNameFormat = "kuberty-%s-%s"
	nameField = "metadata.name"
	defaultServiceAccountName = "default"
)

var (
	faultyNamespaceRegexError ="BadNamespaceFilterError"
	faultyNamespaceRegexMessage = "%v namespaceFilters are not regex."
	resolver = DefaultNamespaceResolver{}
)
type FaultyNamespaceRegexError struct {
	Msg string// description of error
	Err string
}
func (s FaultyNamespaceRegexError) Error() string {
	return s.Err
}

type State struct {
	MasterSecret v1.Secret
	Namespaces   map[string]*NamespaceState // key is the namespace, value is the child state in that namespace
}

// A small struct we use to reduce the amount of parameters in our getActualState and getDesiredState
type Reconciliation struct {
	Client     client.Client
	Context    context.Context
	Registry   v1beta1.Registry
	Namespaces  v1.NamespaceList
}
type NamespaceState struct {
	ChildSecret     v1.Secret         // Not always present
	ServiceAcccount v1.ServiceAccount // Should always be present
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

	//fetch the actual state
	actualState, err := reconciliation.getActualState()
	if err != nil {
		log.Error(err, "Unable to fetch Actual state")
	}
	//fetch the desired state
	desiredState, err := reconciliation.getDesiredState(actualState, resolver)
	if err != nil {
		log.Error(err, "Unable to fetch Actual state")
	}

	//reconcile current and desired
	r.executeReconcile(actualState, desiredState)
	return ctrl.Result{}, nil
}
func (r *KuberdonReconciler) executeReconcile(actualState State, desiredState State) error {
	//masterSecretExists := actualState.MasterSecret.Name != ""

	// diff desired namespace with actual namespace
	return nil
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


func getChildSecretName(controllerName string, secretName string) string {
	return fmt.Sprintf(childSecretNameFormat, controllerName, secretName)
}