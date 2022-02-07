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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	Plan                  string            `json:"plan"`
	Facility              []string          `json:"facility,omitempty"`
	Metro                 string            `json:"metro,omitempty"`
	OS                    string            `json:"operatingSystem"`
	BillingCycle          string            `json:"billingCycle"`
	ProjectID             string            `json:"projectID,omitempty"`
	UserData              string            `json:"userdata,omitempty"`
	Tags                  []string          `json:"tags,omitempty"`
	Description           string            `json:"description,omitempty"`
	IPXEScriptURL         string            `json:"ipxeScriptUrl,omitempty"`
	PublicIPv4SubnetSize  int               `json:"publicIPv4SubnetSize,omitempty"`
	AlwaysPXE             bool              `json:"alwaysPxe,omitempty"`
	HardwareReservationID string            `json:"hardwareReservation_id,omitempty"`
	SpotInstance          bool              `json:"spotInstance,omitempty"`
	SpotPriceMax          resource.Quantity `json:"spotPriceMax,omitempty,string"`
	CustomData            string            `json:"customData,omitempty"`
	UserSSHKeys           []string          `json:"usersshKeys,omitempty"`
	ProjectSSHKeys        []string          `json:"projectsshKeys,omitempty"`
	Features              map[string]string `json:"features,omitempty"`
	NoSSHKeys             bool              `json:"nosshKeys,omitempty"`
	Secret                string            `json:"credentialSecret"`
}

// InstanceStatus defines the observed state of Instance
type InstanceStatus struct {
	Status     string `json:"status"`
	InstanceID string `json:"instanceID"`
	PublicIP   string `json:"publicIP"`
	PrivateIP  string `json:"privateIP"`
	Facility   string `json:"facility"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="InstanceId",type="string",JSONPath=`.status.instanceID`
//+kubebuilder:printcolumn:name="PublicIP",type="string",JSONPath=`.status.publicIP`
//+kubebuilder:printcolumn:name="PrivateIP",type="string",JSONPath=`.status.privateIP`
//+kubebuilder:printcolumn:name="Facility",type="string",JSONPath=`.status.facility`
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.status`

type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}
