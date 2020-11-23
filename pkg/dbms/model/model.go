package model

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

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

//UserEnvironmentRole UserEnvironmentRole
type UserEnvironmentRole struct {
	gorm.Model
	UserID              uint `json:"userId"`
	EnvironmentID       uint `json:"environmentId"`
	SecurityOperationID uint `json:"securityOperationId"`
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

//CopyVariableValue CopyVariableValue
type CopyVariableValue struct {
	SrcVarID uint   `json:"srcVarId"`
	TarEnvID uint   `json:"tarEnvId"`
	TarVarID uint   `json:"tarVarId"`
	NewValue string `json:"newValue"`
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
	EnvironmentIDs   []int            `json:"environmentIds"`
	Deployables      []InstallPayload `json:"deployables"`
}

//RabbitInstallPayload -> Struct of data to post on queue to install
type RabbitInstallPayload struct {
	ProductVersionID int            `json:"productVersionId"`
	Environment      Environment    `json:"environment"`
	Deployable       InstallPayload `json:"deployable"`
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

type selectItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

//CompareEnvironments Model Payload
type CompareEnvironments struct {
	SourceEnvID             int           `json:"sourceEnvId"`
	TargetEnvID             int           `json:"targetEnvId"`
	ExceptCharts            []string      `json:"exceptCharts"`
	OnlyCharts              []string      `json:"onlyCharts"`
	ExceptFields            []string      `json:"exceptFields"`
	OnlyFields              []string      `json:"onlyFields"`
	CustomFields            []FilterField `json:"customFields"`
	FilterOnlyExceptChart   int           `json:"filterOnlyExceptChart"`
	FilterOnlyExceptField   int           `json:"filterOnlyExceptField"`
	SelectedFilterFieldType selectItem    `json:"selectedFilterFieldType"`
	GlobalFilter            string        `json:"globalFilter"`
}

//SaveCompareEnvQuery SaveCompareEnvQuery
type SaveCompareEnvQuery struct {
	ID        uint                `json:"id"`
	Name      string              `json:"name"`
	UserEmail string              `json:"userEmail"`
	Data      CompareEnvironments `json:"data"`
}

//CompareEnvsQuery CompareEnvsQuery
type CompareEnvsQuery struct {
	gorm.Model
	Name   string         `json:"name"`
	UserID int            `json:"userId"`
	Query  postgres.Jsonb `json:"query"`
}

//FilterField Filters the CompareEnvironments result
type FilterField struct {
	FilterType  string `json:"filterType"`
	FilterValue string `json:"filterValue"`
}

//EnvironmentsDiff Model Response
type EnvironmentsDiff struct {
	SourceEnvID int    `json:"sourceEnvId"`
	TargetEnvID int    `json:"targetEnvId"`
	SourceScope string `json:"sourceScope"`
	TargetScope string `json:"targetScope"`
	SourceName  string `json:"sourceName"`
	TargetName  string `json:"targetName"`
	SourceValue string `json:"sourceValue"`
	TargetValue string `json:"targetValue"`
	SourceVarID string `json:"sourceVarId"`
	TargetVarID string `json:"targetVarId"`
}

//CompareEnvsResponse CompareEnvsResponse
type CompareEnvsResponse struct {
	List []EnvironmentsDiff `json:"list"`
}
