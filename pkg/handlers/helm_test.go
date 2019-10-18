package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	mockAud "github.com/softplan/tenkai-api/pkg/audit/mocks"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCharts(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockObject := &mockSvc.HelmServiceInterface{}

	data := make([]model.SearchResult, 1)
	data[0].Name = "test-chart"
	data[0].ChartVersion = "1.0"
	data[0].Description = "Test only"
	data[0].AppVersion = "1.0"

	mockObject.On("SearchCharts", mock.Anything, mock.Anything).Return(&data)
	appContext.HelmServiceAPI = mockObject

	req, err := http.NewRequest("GET", "/listCharts", bytes.NewBuffer(nil))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listCharts)
	handler.ServeHTTP(rr, req)

	mockObject.AssertNumberOfCalls(t, "SearchCharts", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, getExpect(data), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestDeleteHelmRelease(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)
	envs := []model.Environment{env}
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(envs, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("DeleteHelmRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockAudit := mockAud.AuditingInterface{}
	mockAudit.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = &mockAudit

	roles := []string{"tenkai-user"}
	principal := model.Principal{Name: "alfa", Email: "beta", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "DeleteHelmRelease", 1)
	mockAudit.AssertNumberOfCalls(t, "DoAudit", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestRollback(t *testing.T) {

	payloadStr, _ := json.Marshal(getRevision())

	req, err := http.NewRequest("POST", "/rollback", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("RollbackRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.rollback)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "RollbackRelease", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestRevision(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())

	req, err := http.NewRequest("POST", "/revision", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	yaml := "foo: bar"
	mockHelmSvc.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(yaml, nil)

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.revision)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "Get", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, "\"foo: bar\"", string(rr.Body.Bytes()), "Response is not correct.")
}

func getExpect(sr []model.SearchResult) string {
	j, _ := json.Marshal(sr)
	return "{\"charts\":" + string(j) + "}"
}

func getRevision() *model.GetRevisionRequest {
	var revision model.GetRevisionRequest
	revision.EnvironmentID = 999
	revision.ReleaseName = "foo"
	revision.Revision = 800
	return &revision
}
