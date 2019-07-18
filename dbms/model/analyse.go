package model

type DepAnalyse struct {
	Nodes []Node    `json:"nodes"`
	Links []DepLink `json:"links"`
}

type Node struct {
	ID         string `json:"id"`
	Color      string `json:"color"`
	SymbolType string `json:"symbolType"`
	Svg        string `json:"svg"`
}

type DepLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type DepAnalyseRequest struct {
	EnvironmentID int    `json:"environmentId"`
	ChartName     string `json:"chartName"`
	Tag           string `json:"tag"`
}
