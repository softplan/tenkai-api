package model

//Pod structure
type Pod struct {
	Name     string `json:"name"`
	Ready    string `json:"ready"`
	Status   string `json:"status"`
	Restarts int    `json:"restarts"`
	Age      string `json:"age"`
}

//PodResult structure
type PodResult struct {
	Pods []Pod `json:"pods"`
}
