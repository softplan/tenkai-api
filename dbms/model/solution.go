package model

import "github.com/jinzhu/gorm"

//Solution struct
type Solution struct {
	gorm.Model
	Name string `json:"name"`
	Team string `json:"team"`
}

//SolutionResult struct
type SolutionResult struct {
	List []Solution `json:"list"`
}

//SolutionChart struct
type SolutionChart struct {
	gorm.Model
	SolutionID int    `json:"solution_id"`
	ChartName  string `json:"chartName"`
}

//SolutionChartResult struct
type SolutionChartResult struct {
	List []SolutionChart `json:"list"`
}
