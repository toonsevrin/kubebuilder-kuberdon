package controllers

import (
	"fmt"
	"github.com/kuberty/kuberdon/controllers/watchers"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s Reconciliation) getDesiredState(currentState State, resolver NamespaceResolver) (State, error) {
	state := State{}
	// setup desired master secret
	state.MasterSecret = currentState.MasterSecret
	masterSecretExists := state.MasterSecret.Name != ""

	// Setup desired namespaces
	namespaceNames := []string{}
	for _, namespace := range s.Namespaces.Items {
		namespaceNames = append(namespaceNames, namespace.Name)
	}
	filteredNamespaces, nFaultyRegexes := resolver.resolve(namespaceNames, s.Registry.Spec.Namespaces)
	if nFaultyRegexes > 0 {
		return State{}, FaultyNamespaceRegexError{Err: faultyNamespaceRegexError, Msg: fmt.Sprintf(faultyNamespaceRegexMessage, nFaultyRegexes)}
	}
	childSecretName := getChildSecretName(s.Registry.Name, s.Registry.Spec.Secret)
	templatecChildSecret := v1.Secret{ObjectMeta: ctrl.ObjectMeta{Name: childSecretName}}
	if masterSecretExists {
		templatecChildSecret.Type = state.MasterSecret.Type
		templatecChildSecret.Data = state.MasterSecret.Data
		templatecChildSecret.StringData = state.MasterSecret.StringData
		templatecChildSecret.Labels = map[string]string{watchers.OwnedByKuberdonLabel: watchers.OwnedByKuberdonValue}
	}

	for _, namespaceName := range filteredNamespaces {
		childSecret := templatecChildSecret.DeepCopy()
		childSecret.Namespace = namespaceName
		serviceAccount := v1.ServiceAccount{ObjectMeta: ctrl.ObjectMeta{Name: defaultServiceAccountName, Namespace: namespaceName}}
		serviceAccount.ImagePullSecrets = []v1.LocalObjectReference{v1.LocalObjectReference{Name: childSecretName}}
		state.Namespaces[namespaceName] = &NamespaceState{ChildSecret: *childSecret, ServiceAcccount: serviceAccount}
	}

	return state, nil

}
