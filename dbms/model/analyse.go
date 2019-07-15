package model

type DepAnalyse struct {
	Nodes []string  `json:"nodes"`
	Links []DepLink `json:"links"`
}

type DepLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type DepAnalyseRequest struct {
	ChartName string
	Tag       string
}
