package model

type DepAnalyse struct {
	Nodes []Node   `json:"nodes"`
	Links []DepLink `json:"links"`
}

type Node struct {
	ID string `json:"id"`
}

type DepLink struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type DepAnalyseRequest struct {
	ChartName string
	Tag       string
}
