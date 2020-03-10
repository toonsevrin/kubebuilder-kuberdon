/*

MIT License

Copyright (c) 2020 kuberty

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// RegistrySpec defines the desired state of the registry deployment.
type RegistrySpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=`.spec.secret`

	// Secret defines the NamespacedName (eg. namespace/name or name for the default namespace) of the secret to deploy.
	// Todo: Add validation to this field according to https://kubernetes.io/docs/concepts/overview/working-with-objects/names/ for the namespace and name (using regex)
	Secret string `json:"secret"`

	//The namespaces which the secret should be deployed to.
	Namespaces []NamespaceFilter `json:"namespaces"`
}

type NamespaceFilter struct {
	//The (regex) filter for the namespace name
	Name string `json:"name"`
}


// State values:
const (
	// The registry secret is deployed to all relevant service accounts
	SyncedState State = "Synced"
	// The source secret was not found
	secretNotFoundState State = "ErrorSecretNotFound"

)

// Describes the state of the resource
// +kubebuilder:validation:Enum=Synced;ErrorSecretNotFound
type State string

// RegistryStatus defines the observed state of Kuberdon.
type RegistryStatus struct {
	State State

}
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Kuberdon is the Schema for the kuberdons API
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec,omitempty"`
	Status RegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KuberdonList contains a list of Kuberdon
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Registry{}, &RegistryList{})
}
