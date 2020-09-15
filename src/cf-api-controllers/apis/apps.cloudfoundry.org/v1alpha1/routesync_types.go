/*


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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ConditionStatus string

const (
	TrueConditionStatus    ConditionStatus = "True"
	FalseConditionStatus   ConditionStatus = "False"
	UnknownConditionStatus ConditionStatus = "Unknown"
)

// Loosely following this KEP:
// https://github.com/kubernetes/enhancements/tree/master/keps/sig-api-machinery/1623-standardize-conditions
// Eventually we can update to use standard Kubernetes types
type Condition struct {
	Type               string          `json:"type"`
	Status             ConditionStatus `json:"status"`
	LastTransitionTime metav1.Time     `json:"lastTransitionTime"`
	Reason             string          `json:"reason"`
	Message            string          `json:"message"`
}

const (
	SyncedConditionType      = "Synced"
	CompletedConditionReason = "Completed"
	FailedConditionReason    = "Failed"
)

// RouteSyncSpec defines the desired state of RouteSync
type RouteSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PeriodSeconds int32 `json:"period"`
}

// RouteSyncStatus defines the observed state of RouteSync
type RouteSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// RouteSync is the Schema for the routesyncs API
type RouteSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteSyncSpec   `json:"spec,omitempty"`
	Status RouteSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteSyncList contains a list of RouteSync
type RouteSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RouteSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RouteSync{}, &RouteSyncList{})
}
