package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Apployment describes a database.
type Apployment struct {
	metav1.TypeMeta   			`json:",inline"`
	// +optional
	metav1.ObjectMeta 			`json:"metadata,omitempty"`

	Spec 	ApploymentSpec 		`json:"spec"`
	Status  ApploymentStatus 	`json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ApploymentList is a list of Apployment resources
type ApploymentList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Apployment `json:"items"`
}

// ApploymentSpec is the spec 
type ApploymentSpec struct {
	ApploymentName    string  `json:"apployment_name"`
	Replicas          *int32  `json:"replicas"`
}

type ApploymentStatus struct {
	AvailableReplicas  int32 `json:"available_replicas"`
}
