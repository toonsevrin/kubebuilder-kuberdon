package controllers


import (
	"github.com/kuberty/kuberdon/controllers/watchers"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (s Reconciliation) getActualState() (State, error) {
	state := State{}

	secretNamedNamespace := namespacedName(s.Registry.Spec.Secret)
	childSecretName := getChildSecretName(s.Registry.Name, secretNamedNamespace.Name)
	//fetch the master secret
	if err := s.Client.Get(s.Context, secretNamedNamespace, &state.MasterSecret); err != nil {
		if !apierrors.IsNotFound(err) {
			return State{}, err
		}
	}
	//fetch all childSecrets
	var secrets *v1.SecretList
	if err := s.Client.List(s.Context, secrets,
		&client.MatchingLabels{watchers.OwnedByKuberdonLabel: watchers.OwnedByKuberdonValue},
		client.MatchingFields{watchers.OwningControllerField: s.Registry.Name}); err != nil {
		return State{}, err
	}
	//add the secrets to our state
	for _, secret := range secrets.Items {
		if _, ok := state.Namespaces[secret.Namespace]; !ok {
			state.Namespaces[secret.Namespace] = &NamespaceState{ChildSecret: secret}
		}else {
			state.Namespaces[secret.Namespace].ChildSecret = secret
		}

	}

	// Fetch all default service accounts with the childSecretName as an imagePullSecret
	// A potential improvement (if you have many many namespaces) is to add a kuberdon.kuberty.io label to the service account for filtering

	var serviceAccounts *v1.ServiceAccountList
	if err := s.Client.List(s.Context, serviceAccounts, client.MatchingFields{nameField: defaultServiceAccountName}); err != nil {
		return State{}, err
	}

	for _, serviceAccount := range serviceAccounts.Items {
		for _, imagePullSecret := range serviceAccount.ImagePullSecrets {
			if imagePullSecret.Name == childSecretName {
				if _, ok := state.Namespaces[serviceAccount.Namespace]; !ok {
					state.Namespaces[serviceAccount.Namespace] = &NamespaceState{ServiceAcccount: serviceAccount}
				}else {
					state.Namespaces[serviceAccount.Namespace].ServiceAcccount = serviceAccount
				}
			}
		}
	}

	return state, nil
}