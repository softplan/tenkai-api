package model

//Settings struct
type Settings struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//SettingsList struct
type SettingsList struct {
	List []Settings
}

//GetSettingsListRequest struct
type GetSettingsListRequest struct {
	List []string
}
