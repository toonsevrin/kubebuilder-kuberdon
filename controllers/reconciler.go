package controllers

import (
	"context"
	"fmt"
	"github.com/kuberty/kuberdon/controllers/watchers"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

func (r *KuberdonReconciler) executeReconcile(ctx context.Context, actualState State, desiredState State) error {
	type stateComparison struct {
		actualNamespace NamespaceState
		desiredNamespace NamespaceState
	}
	//reconcile namespaces:
	//let's first aggregate all namespaces into comparisons between actual and desired
	allNamespaces := map[string]*stateComparison{}
	for name, namespaceState := range actualState.Namespaces {
		val, exists := allNamespaces[name]
		if !exists {
			val = &stateComparison{actualNamespace: *namespaceState}
			allNamespaces[name] = val
		}else {
			val.actualNamespace = *namespaceState
		}
	}
	for name, namespaceState := range desiredState.Namespaces {
		val, exists := allNamespaces[name]
		if !exists {
			val = &stateComparison{desiredNamespace: *namespaceState}
			allNamespaces[name] = val
		}else {
			val.desiredNamespace = *namespaceState
		}
	}
	for _, stateComp := range allNamespaces {
		err := r.executeNamespaceReconciliation(ctx, stateComp.actualNamespace, stateComp.desiredNamespace)
		if err != nil {
			return fmt.Errorf("error with namespace reconciliation: %v", err)
		}
	}
	return nil
}

func (r *KuberdonReconciler) executeNamespaceReconciliation(ctx context.Context, actualNamespace NamespaceState, desiredNamespace NamespaceState) error{
	// Do nothing in case there is no service account (this usually means the namespace is being deleted)
	if actualNamespace.ServiceAcccount.Name == "" {
		print("sa is nill")
		return nil
	}
	err := r.executeChildSecretReconciliation(ctx, actualNamespace.ChildSecret, desiredNamespace.ChildSecret)
	if err != nil {
		return fmt.Errorf("error with child secret reconciliation: %v", err)
	}

	err = r.executeServiceAccountReconciliation(ctx, desiredNamespace.ChildSecret.Name, actualNamespace.ServiceAcccount, desiredNamespace.ServiceAcccount)
	if err != nil {
		return fmt.Errorf("error with child serviceaccount reconciliation: %v", err)
	}

	return nil
}
// Will this work when the namespace is deleted?
func (r *KuberdonReconciler) executeChildSecretReconciliation(ctx context.Context, actualSecret v1.Secret, desiredSecret v1.Secret) error{
	if desiredSecret.Name == "" {	// Secret should be deleted
		if actualSecret.Name != "" {
			if err := r.Client.Delete(ctx, &actualSecret); err != nil{
				return fmt.Errorf("error deleting childSecret: %v",err)
			}
		}
		return nil
	}

	if actualSecret.Name == "" { // Secret should be created
		if err := r.Client.Create(ctx, &desiredSecret); err != nil {
			return fmt.Errorf("error creating childSecret: %v",err)
		}
		return nil
	}
	if secretNeedsUpdating(actualSecret, desiredSecret){ // Secret should be updated
		actualSecret.Data = desiredSecret.Data
		actualSecret.StringData = desiredSecret.StringData
		if actualSecret.Labels == nil {
			actualSecret.Labels = map[string]string{}
		}
		actualSecret.Labels[watchers.OwnedByKuberdonLabel] = watchers.OwnedByKuberdonValue
		if err := r.Client.Update(ctx, &actualSecret); err != nil {
			return fmt.Errorf("error updating childSecret: %v",err)
		}
	}
	return nil
}

func secretNeedsUpdating(actualSecret v1.Secret, desiredSecret v1.Secret) bool {
	return !reflect.DeepEqual(actualSecret.Data, desiredSecret.Data) ||
	   !reflect.DeepEqual(actualSecret.StringData, desiredSecret.StringData) ||
		actualSecret.Labels[watchers.OwnedByKuberdonLabel] != desiredSecret.Labels[watchers.OwnedByKuberdonLabel]
}

func (r *KuberdonReconciler) executeServiceAccountReconciliation(ctx context.Context, secretName string, actualSA v1.ServiceAccount, desiredSA v1.ServiceAccount) error {
	if actualSA.Name == "" { // ServiceAccount does not exist, namespace is probably being deleted, return
		return nil
	}
	//Cleanup old secrets
	prefix := fmt.Sprintf(childSecretPrefix, secretName)
	newSecrets := []v1.LocalObjectReference{}
	for _, secret := range actualSA.ImagePullSecrets {
		if len(secret.Name) < len(prefix) || secret.Name[0:len(prefix)] != prefix {
			newSecrets = append(newSecrets, secret)
		}else if secret.Name == secretName &&  hasPullSecret(desiredSA, secretName){
			newSecrets = append(newSecrets, secret)
			}
	}
	updated := false
	if reflect.DeepEqual(newSecrets, actualSA.Secrets){
		actualSA = *actualSA.DeepCopy()
		updated = true
		actualSA.ImagePullSecrets = newSecrets
	}
	if !hasPullSecret(actualSA, secretName) && hasPullSecret(desiredSA, secretName){ //add pull secret
		updated = true
		actualSA = *actualSA.DeepCopy()
		actualSA.ImagePullSecrets = append(actualSA.ImagePullSecrets, v1.LocalObjectReference{Name: secretName})

	}
	if updated {
		if err := r.Update(ctx, &actualSA); err != nil {
			return fmt.Errorf("error updating sa imagePullSecrets: %v", err)
		}
	}
	return nil
}

func hasPullSecret(serviceAccount v1.ServiceAccount, secretName string) bool{
	return pullSecretIndex(serviceAccount, secretName) != -1
}

func pullSecretIndex(serviceAccount v1.ServiceAccount, secretName string) int{
	for i, pullSecret := range serviceAccount.ImagePullSecrets {
		if pullSecret.Name == secretName {
			return i
		}
	}
	return -1
}