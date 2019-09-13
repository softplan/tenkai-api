package model

//Pod structure
type Pod struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Ready    string `json:"ready"`
	Status   string `json:"status"`
	Restarts int    `json:"restarts"`
	Age      string `json:"age"`
}

//PodResult structure
type PodResult struct {
	Pods []Pod `json:"pods"`
}

//Service structure
type Service struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	ClusterIP  string `json:"clusterIP"`
	ExternalIP string `json:"externalIP"`
	Ports      string `json:"ports"`
	Age        string `json:"age"`
}

//ServiceResult structure
type ServiceResult struct {
	Services []Service `json:"services"`
}
