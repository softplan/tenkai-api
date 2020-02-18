package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	sVars := mockSourceEnvs()
	tVars := mockTargetVars()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"882","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"user","targetName":"user","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)

	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"888","targetVarId":"998"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"887","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"997"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"992"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"881","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"883","targetVarId":"993"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart3","targetScope":"repo/chart3","sourceName":"foo","targetName":"","sourceValue":"bar","targetValue":"","sourceVarId":"889","targetVarId":""}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterExceptCharts(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	exceptCharts = append(exceptCharts, "global")
	exceptCharts = append(exceptCharts, "repo/chart1")
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
	sVars := mockSourceEnvs()
	tVars := mockTargetVars()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"882","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"user","targetName":"user","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"887","targetVarId":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"888","targetVarId":"998"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"997"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"883","targetVarId":"993"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"881","targetVarId":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"992"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart3","targetScope":"repo/chart3","sourceName":"foo","targetName":"","sourceValue":"bar","targetValue":"","sourceVarId":"889","targetVarId":""}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterOnlyCharts(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	onlyCharts := make([]string, 0)
	onlyCharts = append(onlyCharts, "global")
	onlyCharts = append(onlyCharts, "repo/chart1")

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
	sVars := mockSourceEnvs()
	tVars := mockTargetVars()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"882","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"user","targetName":"user","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)

	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"887","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"888","targetVarId":"998"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"997"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"883","targetVarId":"993"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"881","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"992"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart3","targetScope":"","sourceName":"foo","targetName":"","sourceValue":"bar","targetValue":"","sourceVarId":"889","targetVarId":""}}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterExceptFields(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	onlyCharts := make([]string, 0)
	exceptFields := make([]string, 0)
	exceptFields = append(exceptFields, "f2")
	onlyFields := make([]string, 0)

	var p model.CompareEnvironments
	p.SourceEnvID = 888
	p.TargetEnvID = 999
	p.ExceptCharts = exceptCharts
	p.OnlyCharts = onlyCharts
	p.ExceptFields = exceptFields
	p.OnlyFields = onlyFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvs()
	tVars := mockTargetVars()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"882","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"user","targetName":"user","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)

	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"887","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"888","targetVarId":"998"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"997"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"883","targetVarId":"993"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"881","targetVarId":""}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"992"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart3","targetScope":"repo/chart3","sourceName":"foo","targetName":"","sourceValue":"bar","targetValue":"","sourceVarId":"889","targetVarId":""}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterOnlyFields(t *testing.T) {
	appContext := AppContext{}

	exceptCharts := make([]string, 0)
	onlyCharts := make([]string, 0)
	exceptFields := make([]string, 0)
	onlyFields := make([]string, 0)
	onlyFields = append(onlyFields, "f2")

	var p model.CompareEnvironments
	p.SourceEnvID = 888
	p.TargetEnvID = 999
	p.ExceptCharts = exceptCharts
	p.OnlyCharts = onlyCharts
	p.ExceptFields = exceptFields
	p.OnlyFields = onlyFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvs()
	tVars := mockTargetVars()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"882","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f1","targetName":"f1","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"user","targetName":"user","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"pass","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"887","targetVarId":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"foo","targetName":"foo","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"888","targetVarId":"998"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"global","targetScope":"global","sourceName":"","targetName":"port","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"997"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f4","targetName":"f4","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"883","targetVarId":"993"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"f2","targetName":"","sourceValue":"only-in-source","targetValue":"","sourceVarId":"881","targetVarId":""}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"","targetName":"f3","sourceValue":"","targetValue":"only-in-target","sourceVarId":"","targetVarId":"992"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"f2","targetName":"f2","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart3","targetScope":"","sourceName":"foo","targetName":"","sourceValue":"bar","targetValue":"","sourceVarId":"889","targetVarId":""}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterCustomByStartsWith(t *testing.T) {
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

	var f2 model.FilterField
	f2.FilterType = "StartsWith"
	f2.FilterValue = "urlapi"

	var customFields []model.FilterField
	customFields = append(customFields, f2)

	p.CustomFields = customFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsFilterFields()
	tVars := mockTargetVarsFilterFields()
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
	assert.Contains(t, r, `{"list":[`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService","targetName":"urlapiMyService","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"881","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"istio.virtualservices.enabled","targetName":"istio.virtualservices.enabled","sourceValue":"true","targetValue":"false","sourceVarId":"882","targetVarId":"992"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"image.tag","targetName":"image.tag","sourceValue":"latest","targetValue":"stable","sourceVarId":"883","targetVarId":"993"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService2","targetName":"urlapiMyService2","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field1","targetName":"field1","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field2","targetName":"field2","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterCustomByContains(t *testing.T) {
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

	var f1 model.FilterField
	f1.FilterType = "Contains"
	f1.FilterValue = "virtualservices"

	var customFields []model.FilterField
	customFields = append(customFields, f1)

	p.CustomFields = customFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsFilterFields()
	tVars := mockTargetVarsFilterFields()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService","targetName":"urlapiMyService","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"881","targetVarId":"991"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"istio.virtualservices.enabled","targetName":"istio.virtualservices.enabled","sourceValue":"true","targetValue":"false","sourceVarId":"882","targetVarId":"992"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"image.tag","targetName":"image.tag","sourceValue":"latest","targetValue":"stable","sourceVarId":"883","targetVarId":"993"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService2","targetName":"urlapiMyService2","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field1","targetName":"field1","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field2","targetName":"field2","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterCustomByEndsWith(t *testing.T) {
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

	var f3 model.FilterField
	f3.FilterType = "EndsWith"
	f3.FilterValue = "tag"

	var customFields []model.FilterField
	customFields = append(customFields, f3)

	p.CustomFields = customFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsFilterFields()
	tVars := mockTargetVarsFilterFields()
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
	assert.Contains(t, r, `{"list":[`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService","targetName":"urlapiMyService","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"881","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"istio.virtualservices.enabled","targetName":"istio.virtualservices.enabled","sourceValue":"true","targetValue":"false","sourceVarId":"882","targetVarId":"992"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"image.tag","targetName":"image.tag","sourceValue":"latest","targetValue":"stable","sourceVarId":"883","targetVarId":"993"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService2","targetName":"urlapiMyService2","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field1","targetName":"field1","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field2","targetName":"field2","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterCustomByRegex(t *testing.T) {
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

	var f3 model.FilterField
	f3.FilterType = "RegEx"
	f3.FilterValue = "urlapiMyService|field1"

	var customFields []model.FilterField
	customFields = append(customFields, f3)

	p.CustomFields = customFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsFilterFields()
	tVars := mockTargetVarsFilterFields()
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
	assert.Contains(t, r, `{"list":[`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService","targetName":"urlapiMyService","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"881","targetVarId":"991"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"istio.virtualservices.enabled","targetName":"istio.virtualservices.enabled","sourceValue":"true","targetValue":"false","sourceVarId":"882","targetVarId":"992"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"image.tag","targetName":"image.tag","sourceValue":"latest","targetValue":"stable","sourceVarId":"883","targetVarId":"993"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService2","targetName":"urlapiMyService2","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field1","targetName":"field1","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field2","targetName":"field2","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)
	assert.Contains(t, r, `]}`)
}

func TestCompareEnvironmentsFilterCustom(t *testing.T) {
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

	var f1 model.FilterField
	f1.FilterType = "Contains"
	f1.FilterValue = "virtualservices"

	var f2 model.FilterField
	f2.FilterType = "StartsWith"
	f2.FilterValue = "urlapi"

	var f3 model.FilterField
	f3.FilterType = "EndsWith"
	f3.FilterValue = "tag"

	var customFields []model.FilterField
	customFields = append(customFields, f1)
	customFields = append(customFields, f2)
	customFields = append(customFields, f3)

	p.CustomFields = customFields

	mockVarDao := &mockRepo.VariableDAOInterface{}
	sVars := mockSourceEnvsFilterFields()
	tVars := mockTargetVarsFilterFields()
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
	assert.Contains(t, r, `{"list":[`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService","targetName":"urlapiMyService","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"881","targetVarId":"991"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"istio.virtualservices.enabled","targetName":"istio.virtualservices.enabled","sourceValue":"true","targetValue":"false","sourceVarId":"882","targetVarId":"992"}`)
	assert.Contains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"image.tag","targetName":"image.tag","sourceValue":"latest","targetValue":"stable","sourceVarId":"883","targetVarId":"993"}`)

	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart1","targetScope":"repo/chart1","sourceName":"urlapiMyService2","targetName":"urlapiMyService2","sourceValue":"equal","targetValue":"equal","sourceVarId":"884","targetVarId":"994"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field1","targetName":"field1","sourceValue":"not-equal-1","targetValue":"not-equal-2","sourceVarId":"885","targetVarId":"995"}`)
	assert.NotContains(t, r, `{"sourceEnvId":888,"targetEnvId":999,"sourceScope":"repo/chart2","targetScope":"repo/chart2","sourceName":"field2","targetName":"field2","sourceValue":"equal","targetValue":"equal","sourceVarId":"886","targetVarId":"996"}`)
	assert.Contains(t, r, `]}`)
}

func TestSaveCompareEnvironmentView(t *testing.T) {
	appContext := AppContext{}

	p := mockCompareEnvPayload()

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("SaveCompareEnvsQuery", mock.Anything).Return(0, nil)

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("POST", "/save-compare-env-query", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveCompareEnvQuery)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not 201.")
}

func TestSaveCompareEnvironmentView_UnmarshalError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/save-compare-env-query", appContext.saveCompareEnvQuery)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSaveCompareEnvironmentView_FindByEmailError(t *testing.T) {
	appContext := AppContext{}

	p := mockCompareEnvPayload()

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, errors.New("some error"))

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("SaveCompareEnvsQuery", mock.Anything).Return(0, nil)

	appContext.Repositories.UserDAO = mockUserDao

	req, err := http.NewRequest("POST", "/save-compare-env-query", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveCompareEnvQuery)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSaveCompareEnvironmentView_SaveError(t *testing.T) {
	appContext := AppContext{}

	p := mockCompareEnvPayload()

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("SaveCompareEnvsQuery", mock.Anything).Return(0, errors.New("some errror"))

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("POST", "/save-compare-env-query", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveCompareEnvQuery)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteCompareEnvQuery(t *testing.T) {
	appContext := AppContext{}

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("DeleteCompareEnvQuery", 999).Return(nil)

	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("DELETE", "/compare-environments/delete-query/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/compare-environments/delete-query/{id}", appContext.deleteCompareEnvQuery).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockCompareEnvDao.AssertNumberOfCalls(t, "DeleteCompareEnvQuery", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteCompareEnvQuery_Error(t *testing.T) {
	appContext := AppContext{}

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("DeleteCompareEnvQuery", 999).Return(errors.New("some error"))

	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("DELETE", "/compare-environments/delete-query/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/compare-environments/delete-query/{id}", appContext.deleteCompareEnvQuery).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockCompareEnvDao.AssertNumberOfCalls(t, "DeleteCompareEnvQuery", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestLoadCompareEnvQueries(t *testing.T) {
	appContext := AppContext{}

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	var item model.CompareEnvsQuery
	item.ID = 999
	item.Name = "test"
	item.UserID = 888

	query := json.RawMessage(`{"foo":"bar"}`)
	item.Query = postgres.Jsonb{RawMessage: query}

	var result []model.CompareEnvsQuery
	result = append(result, item)

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("GetByUser", mock.Anything).Return(result, nil)

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("GET", "/compare-environments/load-queries", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.loadCompareEnvQueries)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be 200.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `[{"ID":999,`)
	assert.Contains(t, response, `"name":"test","userId":888,"query":{"foo":"bar"}`)
	assert.Contains(t, response, `}]`)
}

func TestLoadCompareEnvQueries_FindByEmailError(t *testing.T) {
	appContext := AppContext{}

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, errors.New("some error"))

	appContext.Repositories.UserDAO = mockUserDao

	req, err := http.NewRequest("GET", "/compare-environments/load-queries", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.loadCompareEnvQueries)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestLoadCompareEnvQueries_GetByUserError(t *testing.T) {
	appContext := AppContext{}

	var user model.User
	user.ID = 999

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	var item model.CompareEnvsQuery
	item.ID = 999
	item.Name = "test"
	item.UserID = 888

	query := json.RawMessage(`{"foo":"bar"}`)
	item.Query = postgres.Jsonb{RawMessage: query}

	var result []model.CompareEnvsQuery
	result = append(result, item)

	mockCompareEnvDao := &mockRepo.CompareEnvsQueryDAOInterface{}
	mockCompareEnvDao.On("GetByUser", mock.Anything).Return(result, errors.New("some error"))

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.CompareEnvsQueryDAO = mockCompareEnvDao

	req, err := http.NewRequest("GET", "/compare-environments/load-queries", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.loadCompareEnvQueries)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func mockVar(id int, envID int, repo string, field string, value string) model.Variable {
	var v model.Variable
	v.ID = uint(id)
	v.Scope = repo
	v.Name = field
	v.Value = value
	v.Secret = false
	v.EnvironmentID = envID
	return v
}

func mockSourceEnvs() []model.Variable {
	var sVars []model.Variable
	// Source chart 1
	sVars = append(sVars, mockVar(881, 888, "repo/chart1", "f2", "only-in-source"))
	sVars = append(sVars, mockVar(882, 888, "repo/chart1", "f1", "equal"))
	sVars = append(sVars, mockVar(883, 888, "repo/chart1", "f4", "not-equal-1"))
	// Source chart 2
	sVars = append(sVars, mockVar(884, 888, "repo/chart2", "f1", "equal"))
	sVars = append(sVars, mockVar(885, 888, "repo/chart2", "f2", "not-equal-1"))
	// Source global
	sVars = append(sVars, mockVar(886, 888, "global", "user", "equal"))
	sVars = append(sVars, mockVar(887, 888, "global", "pass", "only-in-source"))
	sVars = append(sVars, mockVar(888, 888, "global", "foo", "not-equal-1"))
	// Source chart 3
	sVars = append(sVars, mockVar(889, 888, "repo/chart3", "foo", "bar"))
	return sVars
}

func mockTargetVars() []model.Variable {
	var tVars []model.Variable
	// Target chart 1
	tVars = append(tVars, mockVar(991, 999, "repo/chart1", "f1", "equal"))
	tVars = append(tVars, mockVar(992, 999, "repo/chart1", "f3", "only-in-target"))
	tVars = append(tVars, mockVar(993, 999, "repo/chart1", "f4", "not-equal-2"))
	// Target chart 2
	tVars = append(tVars, mockVar(994, 999, "repo/chart2", "f1", "equal"))
	tVars = append(tVars, mockVar(995, 999, "repo/chart2", "f2", "not-equal-2"))
	// Target global
	tVars = append(tVars, mockVar(996, 999, "global", "user", "equal"))
	tVars = append(tVars, mockVar(997, 999, "global", "port", "only-in-target"))
	tVars = append(tVars, mockVar(998, 999, "global", "foo", "not-equal-2"))
	return tVars
}

func mockSourceEnvsFilterFields() []model.Variable {
	var sVars []model.Variable
	// Source chart 1
	sVars = append(sVars, mockVar(881, 888, "repo/chart1", "urlapiMyService", "not-equal-1"))
	sVars = append(sVars, mockVar(882, 888, "repo/chart1", "istio.virtualservices.enabled", "true"))
	sVars = append(sVars, mockVar(883, 888, "repo/chart1", "image.tag", "latest"))
	sVars = append(sVars, mockVar(884, 888, "repo/chart1", "urlapiMyService2", "equal"))
	// Source chart 2
	sVars = append(sVars, mockVar(885, 888, "repo/chart2", "field1", "not-equal-1"))
	sVars = append(sVars, mockVar(886, 888, "repo/chart2", "field2", "equal"))
	return sVars
}

func mockTargetVarsFilterFields() []model.Variable {
	var tVars []model.Variable
	// Target chart 1
	tVars = append(tVars, mockVar(991, 999, "repo/chart1", "urlapiMyService", "not-equal-2"))
	tVars = append(tVars, mockVar(992, 999, "repo/chart1", "istio.virtualservices.enabled", "false"))
	tVars = append(tVars, mockVar(993, 999, "repo/chart1", "image.tag", "stable"))
	tVars = append(tVars, mockVar(994, 999, "repo/chart1", "urlapiMyService2", "equal"))
	// Target chart 2
	tVars = append(tVars, mockVar(995, 999, "repo/chart2", "field1", "not-equal-2"))
	tVars = append(tVars, mockVar(996, 999, "repo/chart2", "field2", "equal"))
	return tVars
}

func mockCompareEnvPayload() model.SaveCompareEnvQuery {

	exceptCharts := make([]string, 0)
	onlyCharts := make([]string, 0)
	exceptFields := make([]string, 0)
	onlyFields := make([]string, 0)

	var data model.CompareEnvironments
	data.SourceEnvID = 888
	data.TargetEnvID = 999
	data.ExceptCharts = exceptCharts
	data.OnlyCharts = onlyCharts
	data.ExceptFields = exceptFields
	data.OnlyFields = onlyFields

	var f1 model.FilterField
	f1.FilterType = "Contains"
	f1.FilterValue = "virtualservices"

	var f2 model.FilterField
	f2.FilterType = "StartsWith"
	f2.FilterValue = "urlapi"

	var f3 model.FilterField
	f3.FilterType = "EndsWith"
	f3.FilterValue = "tag"

	var customFields []model.FilterField
	customFields = append(customFields, f1)
	customFields = append(customFields, f2)
	customFields = append(customFields, f3)

	data.CustomFields = customFields

	var p model.SaveCompareEnvQuery
	p.Name = "my query"
	p.UserEmail = "foo@bar.com"
	p.Data = data

	return p
}
