package handlers

import (
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func TestSplitSrvNameIfNeeded(t *testing.T) {
	assert.Equal(t, "repo/my-chart", SplitSrvNameIfNeeded("repo/my-chart - 0.1.0"))
	assert.Equal(t, "repo/my-chart", SplitSrvNameIfNeeded("repo/my-chart"))
}

func TestSplitChartVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", SplitChartVersion("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", SplitChartVersion("repo/my-chart"))
}

func TestSplitChartRepo(t *testing.T) {
	assert.Equal(t, "repo", SplitChartRepo("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", SplitChartRepo("my-chart"))
}

func TestGetChartLatestVersion(t *testing.T) {
	appContext := AppContext{}

	var sr1 model.SearchResult
	sr1.Name = "repo/my-chart"
	sr1.ChartVersion = "0.1.0"
	sr1.AppVersion = "1.0.0"
	sr1.Description = "This is my chart"

	var results []model.SearchResult
	results = append(results, sr1)

	latestVersion := appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "", latestVersion, "Should not have a latest version")

	var sr2 model.SearchResult
	sr2.Name = "repo/my-chart"
	sr2.ChartVersion = "0.2.0"
	sr2.AppVersion = "1.0.0"
	sr2.Description = "This is my chart"
	results = append(results, sr2)

	latestVersion = appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "0.2.0", latestVersion, "Latest version should be 0.2.0")
}
