package model

//HistoryRequest Structure
type HistoryRequest struct {
	EnvironmentID int    `json:"environmentID"`
	ReleaseName   string `json:"releaseName"`
}

//GetRevisionRequest Structure
type GetRevisionRequest struct {
	EnvironmentID int    `json:"environmentID"`
	ReleaseName   string `json:"releaseName"`
	Revision      int    `json:"revision"`
}
