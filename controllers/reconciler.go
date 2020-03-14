package controllers

func (r *KuberdonReconciler) executeReconcile(actualState State, desiredState State) error {
	masterSecretExists := actualState.MasterSecret.Name != ""

	// diff desired namespace with actual namespace
	return nil
}