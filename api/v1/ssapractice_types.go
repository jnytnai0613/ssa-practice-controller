/*
Copyright 2022.

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

package v1

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
)

type DeploymentSpecApplyConfiguration appsv1apply.DeploymentSpecApplyConfiguration

func (c *DeploymentSpecApplyConfiguration) DeepCopy() *DeploymentSpecApplyConfiguration {
	out := new(DeploymentSpecApplyConfiguration)
	bytes, err := json.Marshal(c)
	if err != nil {
		panic("Failed to marshal")
	}
	err = json.Unmarshal(bytes, out)
	if err != nil {
		panic("Failed to unmarshal")
	}
	return out
}

// SSAPracticeSpec defines the desired state of SSAPractice
type SSAPracticeSpec struct {
	DepSpec *DeploymentSpecApplyConfiguration `json:"depSpec"`
}

// SSAPracticeStatus defines the observed state of SSAPractice
type SSAPracticeStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SSAPractice is the Schema for the ssapractices API
type SSAPractice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSAPracticeSpec   `json:"spec,omitempty"`
	Status SSAPracticeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSAPracticeList contains a list of SSAPractice
type SSAPracticeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSAPractice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSAPractice{}, &SSAPracticeList{})
}
