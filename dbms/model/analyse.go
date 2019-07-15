package model

type DepAnalyseResponse struct {
	analyses []DepAnalyse
}

type DepAnalyse struct {
	nodes []string
	links []DepLink
}

type DepLink struct {
	source string
	target string
}

type DepAnalyseRequest struct {
	chartName string
	tag string
}

