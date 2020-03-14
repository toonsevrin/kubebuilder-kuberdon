package controllers

import (
	"fmt"
	"github.com/kuberty/kuberdon/controllers/watchers"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s Reconciliation) getActualState() (State, error) {
	state := State{Namespaces:map[string]*NamespaceState{}}

	secretNamedNamespace := namespacedName(s.Registry.Spec.Secret)
	//fetch the master secret
	if err := s.Client.Get(s.Context, secretNamedNamespace, &state.MasterSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return State{}, fmt.Errorf("Failed to fetch master secret: %v", err)
		}
	}

	//fetch all childSecrets
	var secrets v1.SecretList
	if err := s.Client.List(s.Context, &secrets,
		client.MatchingLabels{watchers.OwnedByKuberdonLabel: watchers.OwnedByKuberdonValue},
		client.MatchingFields{watchers.OwningControllerField: s.Registry.Name}); err != nil {
		return State{}, fmt.Errorf("Failed to fetch child secrets: %v", err)
	}
	//add the secrets to our state, issue with multiple secrets owned by the controller!
	for _, secret := range secrets.Items {
		if _, ok := state.Namespaces[secret.Namespace]; !ok {
			state.Namespaces[secret.Namespace] = &NamespaceState{ChildSecret: secret}
		} else {
			state.Namespaces[secret.Namespace].ChildSecret = secret
		}

	}

	// Fetch all default service accounts

	var serviceAccounts v1.ServiceAccountList
	if err := s.Client.List(s.Context, &serviceAccounts, client.MatchingFields{nameField: defaultServiceAccountName}); err != nil {
		return State{}, fmt.Errorf("Failed to fetch service accounts: %v", err)
	}

	for _, serviceAccount := range serviceAccounts.Items {
				if _, ok := state.Namespaces[serviceAccount.Namespace]; !ok {
					state.Namespaces[serviceAccount.Namespace] = &NamespaceState{ServiceAcccount: serviceAccount}
				} else {
					state.Namespaces[serviceAccount.Namespace].ServiceAcccount = serviceAccount
				}
	}

	return state, nil
}
