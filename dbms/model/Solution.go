package model

import "github.com/jinzhu/gorm"

type Solution struct {
	gorm.Model
	Name string `json:"name"`
	Team string `json:"team"`
}

type SolutionResult struct {
	List []Solution `json:"list"`
}

type SolutionChart struct {
	gorm.Model
	SolutionID int    `json:"solution_id"`
	ChartName  string `json:"chartName"`
}
type SolutionChartResult struct {
	List []SolutionChart `json:"list"`
}
