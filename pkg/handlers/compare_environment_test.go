package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCompareEnvironments(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	onlyCharts := make([]string, 0)
	exceptFields := make([]string, 0)
	onlyFields := make([]string, 0)

	var p model.CompareEnvironments
	p.SourceEnvID = 888
	p.TargetEnvID = 999
	p.ExceptCharts = exceptCharts
	p.OnlyCharts = onlyCharts
	p.ExceptFields = exceptFields
	p.OnlyFields = onlyFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsScenario1()
	tVars := mockTargetVarsScenario1()
	mockVarDao.On("GetAllVariablesByEnvironment", p.SourceEnvID).Return(sVars, nil)
	mockVarDao.On("GetAllVariablesByEnvironment", p.TargetEnvID).Return(tVars, nil)
	appContext.Repositories.VariableDAO = mockVarDao

	req, err := http.NewRequest("POST", "/compare-environments", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.compareEnvironments)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")

	r := string(rr.Body.Bytes())
	fmt.Println(r)
	assert.Contains(t, r, `{"list":[`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterExceptCharts(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	exceptCharts = append(exceptCharts, "global")
	exceptCharts = append(exceptCharts, "repo/chart1")
	onlyCharts := make([]string, 0)

	exceptFields := make([]string, 0)
	// exceptFields = append(exceptFields, "f5")
	// exceptFields = append(exceptFields, "passwords")
	onlyFields := make([]string, 0)

	var p model.CompareEnvironments
	p.SourceEnvID = 888
	p.TargetEnvID = 999
	p.ExceptCharts = exceptCharts
	p.OnlyCharts = onlyCharts
	p.ExceptFields = exceptFields
	p.OnlyFields = onlyFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsScenario1()
	tVars := mockTargetVarsScenario1()
	mockVarDao.On("GetAllVariablesByEnvironment", p.SourceEnvID).Return(sVars, nil)
	mockVarDao.On("GetAllVariablesByEnvironment", p.TargetEnvID).Return(tVars, nil)
	appContext.Repositories.VariableDAO = mockVarDao

	req, err := http.NewRequest("POST", "/compare-environments", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.compareEnvironments)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")

	r := string(rr.Body.Bytes())
	fmt.Println(r)
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2"}`)
	assert.Contains(t, r, `]}`)
}

func mockVar(envID int, repo string, field string, value string) model.Variable {
	var v model.Variable
	v.Scope = repo
	v.Name = field
	v.Value = value
	v.Secret = false
	v.EnvironmentID = envID
	return v
}

func mockSourceEnvsScenario1() []model.Variable {
	var sVars []model.Variable
	// Source chart 1
	sVars = append(sVars, mockVar(888, "repo/chart1", "f1", "equal"))
	sVars = append(sVars, mockVar(888, "repo/chart1", "f2", "only-in-source"))
	sVars = append(sVars, mockVar(888, "repo/chart1", "f4", "not-equal-1"))
	// Source chart 2
	sVars = append(sVars, mockVar(888, "repo/chart2", "f1", "equal"))
	sVars = append(sVars, mockVar(888, "repo/chart2", "f2", "not-equal-1"))
	// sVars = append(sVars, mockVar(888, "repo/chart2", "f5", "ignore"))
	// Source global
	sVars = append(sVars, mockVar(888, "global", "user", "equal"))
	sVars = append(sVars, mockVar(888, "global", "pass", "only-in-source"))
	sVars = append(sVars, mockVar(888, "global", "foo", "not-equal-1"))
	return sVars
}

func mockTargetVarsScenario1() []model.Variable {
	var tVars []model.Variable
	// Target chart 1
	tVars = append(tVars, mockVar(999, "repo/chart1", "f1", "equal"))
	tVars = append(tVars, mockVar(999, "repo/chart1", "f3", "only-in-target"))
	tVars = append(tVars, mockVar(999, "repo/chart1", "f4", "not-equal-2"))
	// Target chart 2
	tVars = append(tVars, mockVar(999, "repo/chart2", "f1", "equal"))
	tVars = append(tVars, mockVar(999, "repo/chart2", "f2", "not-equal-2"))
	// tVars = append(tVars, mockVar(999, "repo/chart2", "f5", "ignored"))
	// Target global
	tVars = append(tVars, mockVar(999, "global", "user", "equal"))
	tVars = append(tVars, mockVar(999, "global", "port", "only-in-target"))
	tVars = append(tVars, mockVar(999, "global", "foo", "not-equal-2"))
	return tVars
}
