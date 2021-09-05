/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/packethost/packngo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ImportKeyPairSpec defines the desired state of ImportKeyPair
type ImportKeyPairSpec struct {
	packngo.SSHKeyCreateRequest
	Secret string `json:"secret"`
}

// ImportKeyPairStatus defines the observed state of ImportKeyPair
type ImportKeyPairStatus struct {
	Status    string `json:"status"`
	KeyPairID string `json:"keyPairID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ImportKeyPair is the Schema for the importkeypairs API
type ImportKeyPair struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImportKeyPairSpec   `json:"spec,omitempty"`
	Status ImportKeyPairStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImportKeyPairList contains a list of ImportKeyPair
type ImportKeyPairList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ImportKeyPair `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ImportKeyPair{}, &ImportKeyPairList{})
}
