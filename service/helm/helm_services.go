package helmapi

import "github.com/softplan/tenkai-api/dbms/model"

//HelmServiceInterface - Interface
type HelmServiceInterface interface {
	SearchCharts(searchTerms []string, allVersions bool) *[]model.SearchResult
}

//HelmServiceImpl - Concrete type
type HelmServiceImpl struct {
}
