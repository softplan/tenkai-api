package model

//HistoryRequest Structure
type HistoryRequest struct {
	EnvironmentID int    `json:"environmentID"`
	ReleaseName   string `json:"releaseName"`
}
