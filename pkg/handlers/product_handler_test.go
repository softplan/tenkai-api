package handlers

import (
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func TestSplitSrvNameIfNeeded(t *testing.T) {
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart - 0.1.0"))
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart"))
}

func TestSplitChartVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", splitChartVersion("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartVersion("repo/my-chart"))
}

func TestSplitChartRepo(t *testing.T) {
	assert.Equal(t, "repo", splitChartRepo("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartRepo("my-chart"))
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

func Test_getNumberOfTag(t *testing.T) {
	assert.Equal(t, uint64(19030015000000), getNumberOfTag("19.3.0-15"))
	assert.Equal(t, uint64(20401025000000), getNumberOfTag("20.40.10-25"))
	assert.Equal(t, uint64(10000000000), getNumberOfTag("0.1.0-0"))
}
