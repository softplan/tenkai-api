package model

import "github.com/jinzhu/gorm"

//Environment - Environment Model
type Environment struct {
	gorm.Model
	Group          string `json:"group"`
	Name           string `json:"name"`
	ClusterURI     string `json:"cluster_uri"`
	CACertificate  string `json:"ca_certificate"`
	Token          string `json:"token"`
	Namespace      string `json:"namespace"`
	Gateway        string `json:"gateway"`
	ProductVersion string `json:"productVersion"`
	CurrentRelease string `json:"currentRelease"`
}

//EnvResult Model
type EnvResult struct {
	Envs []Environment
}

//User struct
type User struct {
	gorm.Model
	Email                string        `json:"email"`
	DefaultEnvironmentID int           `json:"defaultEnvironmentID"`
	Environments         []Environment `gorm:"many2many:user_environment;"`
}

//UserResult struct
type UserResult struct {
	Users []User `json:"users"`
}

//DataElement dataElement
type DataElement struct {
	Data Environment `json:"data"`
}

//DataVariableElement dataElement
type DataVariableElement struct {
	Data Variable `json:"data"`
}

//VariablesResult Model
type VariablesResult struct {
	Variables []Variable
}

//SearchResult result
type SearchResult struct {
	Name         string `json:"name"`
	ChartVersion string `json:"chartVersion"`
	AppVersion   string `json:"appVersion"`
	Description  string `json:"description"`
}

//ChartsResult Model
type ChartsResult struct {
	Charts []SearchResult `json:"charts"`
}

//Variable Structure
type Variable struct {
	gorm.Model
	Scope         string `json:"scope" gorm:"index:var_scope"`
	ChartVersion  string `gorm:"-" json:"chartVersion"`
	Name          string `json:"name" gorm:"index:var_name"`
	Value         string `json:"value"`
	Secret        bool   `json:"secret"`
	Description   string `json:"description"`
	EnvironmentID int    `json:"environmentId"`
}

//VariableData Struct
type VariableData struct {
	Data []Variable `json:"data"`
}

//InstallArguments Method
type InstallArguments struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

//InstallPayload Struct
type InstallPayload struct {
	EnvironmentID int    `json:"environmentId"`
	Chart         string `json:"chart"`
	ChartVersion  string `json:"chartVersion"`
	Name          string `json:"name"`
}

//MultipleInstallPayload struct
type MultipleInstallPayload struct {
	ProductVersionID int              `json:"productVersionId"`
	EnvironmentID    int              `json:"environmentId"`
	Deployables      []InstallPayload `json:"deployables"`
}

//GetChartRequest struct
type GetChartRequest struct {
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
}

//InvalidVariablesResult Model
type InvalidVariablesResult struct {
	InvalidVariables []InvalidVariable
}

//InvalidVariable Model
type InvalidVariable struct {
	Scope        string `json:"scope"`
	Name         string `json:"name"`
	Value        string `json:"value"`
	VariableRule string `json:"variableRule"`
	RuleType     string `json:"ruleType"`
	ValueRule    string `json:"valueRule"`
}
