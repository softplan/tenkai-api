package model

type HistoryRequest struct {
	EnvironmentID int    `json:"environmentID"`
	ReleaseName   string `json:"releaseName"`
}
