package controllers

import (
	"fmt"
	"github.com/kuberty/kuberdon/controllers/watchers"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (s Reconciliation) getDesiredState(currentState State, resolver NamespaceResolver) (State, error) {
	state := State{ Namespaces:map[string]*NamespaceState{}}
	// setup desired master secret
	state.MasterSecret = currentState.MasterSecret
	fmt.Printf("Master sercret: %v\n", currentState.MasterSecret.Name)
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
	childSecretName := getChildSecretName(s.Registry.Name, namespacedName(s.Registry.Spec.Secret).Name)
	templateChildSecret := v1.Secret{ObjectMeta: ctrl.ObjectMeta{}}
	masterSecretCopy := state.MasterSecret.DeepCopy()
	if masterSecretExists {
		templateChildSecret.Name = childSecretName
		templateChildSecret.Type = masterSecretCopy.Type
		templateChildSecret.Data = masterSecretCopy.Data
		templateChildSecret.StringData = masterSecretCopy.StringData
		templateChildSecret.Labels = map[string]string{watchers.OwnedByKuberdonLabel: watchers.OwnedByKuberdonValue}
		controller, blockOwnerDeletion := true, false
		templateChildSecret.OwnerReferences = []v12.OwnerReference{{APIVersion: s.Registry.APIVersion, Kind: s.Registry.Kind, Name: s.Registry.Name, UID: s.Registry.UID, Controller: &controller, BlockOwnerDeletion: &blockOwnerDeletion}}
	}

	for _, namespaceName := range filteredNamespaces {
		childSecret := templateChildSecret.DeepCopy()
		childSecret.Namespace = namespaceName
		serviceAccount := v1.ServiceAccount{ObjectMeta: ctrl.ObjectMeta{Name: defaultServiceAccountName, Namespace: namespaceName}}
		serviceAccount.ImagePullSecrets = []v1.LocalObjectReference{v1.LocalObjectReference{Name: childSecretName}}
		state.Namespaces[namespaceName] = &NamespaceState{ChildSecret: *childSecret, ServiceAcccount: serviceAccount}
	}

	return state, nil

}
