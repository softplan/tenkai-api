package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/gorilla/mux"
	mockAud "github.com/softplan/tenkai-api/pkg/audit/mocks"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/softplan/tenkai-api/pkg/service/core/mocks"
	helmapi "github.com/softplan/tenkai-api/pkg/service/helm"
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

	mockAudit := &mockAud.AuditingInterface{}
	mockAudit.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = mockAudit
	appContext.ConventionInterface = mockConvention

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
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
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

	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.ConventionInterface = mockConvention

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.rollback)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "RollbackRelease", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

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

	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.ConventionInterface = mockConvention

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.revision)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "Get", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, "\"foo: bar\"", string(rr.Body.Bytes()), "Response is not correct.")
}

func TestListReleaseHistory(t *testing.T) {
	var payload model.HistoryRequest
	payload.EnvironmentID = 999
	payload.ReleaseName = "foo"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/listReleaseHistory", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)

	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")

	var info helmapi.ReleaseInfo
	info.Revision = 987
	info.Status = "DEPLOYED"
	info.Chart = "my-helm-chart"
	info.Description = "Install completed"

	var history helmapi.ReleaseHistory
	history = append(history, info)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetHelmReleaseHistory", mock.Anything, mock.Anything).Return(history, nil)

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.ConventionInterface = mockConvention

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listReleaseHistory)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "GetHelmReleaseHistory", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, getExpecHistory(), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestListHelmDeploymentsByEnvironment(t *testing.T) {

	req, err := http.NewRequest("GET", "/listHelmDeploymentsByEnvironment/999", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)

	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")

	var listReleases []helmapi.ListRelease

	lr := helmapi.ListRelease{
		Name:       "my-foo",
		Revision:   9999,
		Updated:    "",
		Status:     "",
		Chart:      "foo",
		AppVersion: "0.1.0",
		Namespace:  "xpto",
	}
	listReleases = append(listReleases, lr)

	result := &helmapi.HelmListResult{
		Next:     "998",
		Releases: listReleases,
	}

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("ListHelmDeployments", mock.Anything, mock.Anything).Return(result, nil)

	appContext := AppContext{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.ConventionInterface = mockConvention
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	j, _ := json.Marshal(result)
	assert.Equal(t, string(j), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestHasConfigMap(t *testing.T) {
	var payload model.GetChartRequest
	payload.ChartName = "foo"
	payload.ChartVersion = "0.1.0"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/hasConfigMap", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}

	partialDeploymentYaml := "...name: {{ template \"foo.name\" . }}-gcm-{{ .Release.Namespace }}..."
	result := []byte(partialDeploymentYaml)
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(result, nil)

	appContext := AppContext{}
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.hasConfigMap)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "GetTemplate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, "{\"result\":\"true\"}", string(rr.Body.Bytes()), "Response is not correct.")
}

func getExpect(sr []model.SearchResult) string {
	j, _ := json.Marshal(sr)
	return "{\"charts\":" + string(j) + "}"
}

func getExpecHistory() string {
	return "[{\"revision\":987,\"updated\":\"\",\"status\":\"DEPLOYED\",\"chart\":\"my-helm-chart\",\"description\":\"Install completed\"}]"
}

func getRevision() *model.GetRevisionRequest {
	var revision model.GetRevisionRequest
	revision.EnvironmentID = 999
	revision.ReleaseName = "foo"
	revision.Revision = 800
	return &revision
}
