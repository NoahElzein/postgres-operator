package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Postgres struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              PostgresSpec   `json:"spec"`
	Status            PostgresStatus `json:"status,omitempty"`
}

type PostgresSpec struct {
	DeploymentName string `json:"deploymentName"`
    Image string `json:"image"`
    Username string `json:"username"`
    Password string `json:"password"`
    Database string `json:"database"`
    Replicas       *int32 `json:"replicas"`
    Commands []string `json:"commands"`
}
type PostgresStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
    ActionHistory []string `json:"actionHistory"`
    VerifyCmd string `json:"verifyCommand"`
    ServiceIP string `json:"serviceIP"`
    ServicePort string `json:"servicePort"`
    Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PostgresList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Postgres `json:"items"`
}
