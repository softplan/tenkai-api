package model

//Chart struct
type Chart struct {
	Repo    string `json:"repo"`
	Name    string `json:"chartName"`
	Version string `json:"chartVersion"`
}

//VariablesByChartAndEnvironment struct
type VariablesByChartAndEnvironment struct {
	EnvironmentID int
	Chart         string
	Variables     []Variable
}

//VariablesDefault struct
type VariablesDefault struct {
	Chart     string
	Variables map[string]interface{}
}

//NewVariable Structure
type NewVariable struct {
	Scope         string `json:"scope"`
	ChartVersion  string `json:"chartVersion"`
	Name          string `json:"name"`
	Value         string `json:"value"`
	Secret        bool   `json:"secret"`
	Description   string `json:"description"`
	EnvironmentID int    `json:"environmentId"`
	New           bool   `json:"new"`
}
