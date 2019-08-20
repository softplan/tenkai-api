package model

//ReleaseDeleteRequest Structure
type ReleaseDeleteRequest struct {
	EnvironmentID int    `json:"environmentID"`
	ReleaseName   string `json:"releaseName"`
	Purge         bool   `json:"purge"`
}
