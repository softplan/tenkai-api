package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockRabbit "github.com/softplan/tenkai-api/pkg/rabbitmq/mocks"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getCharts() *[]model.SearchResult {
	charts := make([]model.SearchResult, 0)
	searchResult := model.SearchResult{}
	searchResult.Name = "foo"
	searchResult.Description = "alfaChart"
	searchResult.ChartVersion = "1.0"
	charts = append(charts, searchResult)
	return &charts
}

func TestListCharts(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockHelmSvc := mockHelmSearchCharts(&appContext)

	req, err := http.NewRequest("GET", "/listCharts", bytes.NewBuffer(nil))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listCharts)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "SearchCharts", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, getExpect(getHelmSearchResult()), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestDeleteHelmRelease(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("DeleteHelmRelease", "./config/foo_bar", "foo", false).Return(nil)

	auditValues := make(map[string]string)
	auditValues["environment"] = "bar"
	auditValues["purge"] = "false"
	auditValues["name"] = "foo"
	mockAudit := mockDoAudit(&appContext, "deleteHelmRelease", auditValues)

	mockConvention := mockConventionInterface(&appContext)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc

	mockPrincipal(req)

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

func TestDeleteHelmRelease_EnvironmentIDQueryError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?xxxxYYYY=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteHelmRelease_ReleaseNameQueryError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&xxxxYYYY=foo&purge=false", nil)
	assert.NoError(t, err)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteHelmRelease_PurgeQueryError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&xxxxYYYY=false", nil)
	assert.NoError(t, err)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteHelmRelease_HasAccessError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(nil, errors.New("Record not found"))

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestDeleteHelmRelease_GetByIDError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByIDError(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteHelmRelease_DeleteHelmReleaseError(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.NoError(t, err)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("DeleteHelmRelease", "./config/foo_bar", "foo", false).Return(errors.New("some error"))

	auditValues := make(map[string]string)
	auditValues["environment"] = "bar"
	auditValues["purge"] = "false"
	auditValues["name"] = "foo"

	mockConvention := mockConventionInterface(&appContext)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "DeleteHelmRelease", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestRollback(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())
	req, err := http.NewRequest("POST", "/rollback", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("RollbackRelease", "./config/foo_bar", "foo", 800).Return(nil)

	mockConvention := mockConventionInterface(&appContext)

	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.rollback)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "RollbackRelease", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestRollback_GetByIDError(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())
	req, err := http.NewRequest("POST", "/rollback", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByIDError(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.rollback)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestRollback_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/rollback", appContext.rollback)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestRollback_RollbackReleaseError(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())
	req, err := http.NewRequest("POST", "/rollback", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("RollbackRelease", "./config/foo_bar", "foo", 800).Return(errors.New("some error"))

	mockConvention := mockConventionInterface(&appContext)

	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.rollback)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "RollbackRelease", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestRevision(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())
	req, err := http.NewRequest("POST", "/revision", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	yaml := "foo: bar"
	mockHelmSvc.On("Get", "./config/foo_bar", "foo", 800).Return(yaml, nil)

	mockConvention := mockConventionInterface(&appContext)

	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.revision)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "Get", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, "\"foo: bar\"", string(rr.Body.Bytes()), "Response is not correct.")
}

func TestRevision_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/revision", appContext.revision)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestRevision_GetByIDError(t *testing.T) {
	payloadStr, _ := json.Marshal(getRevision())
	req, err := http.NewRequest("POST", "/revision", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByIDError(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.revision)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListReleaseHistory(t *testing.T) {
	var payload model.HistoryRequest
	payload.EnvironmentID = 999
	payload.ReleaseName = "foo"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/listReleaseHistory", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockConvention := mockConventionInterface(&appContext)

	var info helmapi.ReleaseInfo
	info.Revision = 987
	info.Status = "DEPLOYED"
	info.Chart = "my-helm-chart"
	info.Description = "Install completed"

	var history helmapi.ReleaseHistory
	history = append(history, info)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetHelmReleaseHistory", "./config/foo_bar", "foo").Return(history, nil)

	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listReleaseHistory)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "GetHelmReleaseHistory", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, getExpecHistory(), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestListReleaseHistory_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/listReleaseHistory", appContext.listReleaseHistory)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListReleaseHistory_GetByIDError(t *testing.T) {
	var payload model.HistoryRequest
	payload.EnvironmentID = 999
	payload.ReleaseName = "foo"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/listReleaseHistory", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByIDError(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listReleaseHistory)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListHelmDeploymentsByEnvironment(t *testing.T) {

	req, err := http.NewRequest("GET", "/listHelmDeploymentsByEnvironment/999", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockListHelmDeployments(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "ListHelmDeployments", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	j, _ := json.Marshal(mockHelmListResult())
	assert.Equal(t, string(j), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestListHelmDeploymentsByEnvironment_StringConvError(t *testing.T) {

	req, err := http.NewRequest("GET", "/listHelmDeploymentsByEnvironment/qwert", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListHelmDeploymentsByEnvironment_GetByIDError(t *testing.T) {

	req, err := http.NewRequest("GET", "/listHelmDeploymentsByEnvironment/999", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByIDError(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListHelmDeploymentsByEnvironment_ListHelmDeploymentsError(t *testing.T) {

	req, err := http.NewRequest("GET", "/listHelmDeploymentsByEnvironment/999", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockListHelmDeploymentsError(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "ListHelmDeployments", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestHasConfigMap(t *testing.T) {
	req, err := http.NewRequest("POST", "/hasConfigMap", getPayloadChartRequest())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}

	partialDeploymentYaml := "...name: {{ template \"my-chart.name\" . }}-gcm-{{ .Release.Namespace }}..."
	result := []byte(partialDeploymentYaml)
	mockHelmSvc.On("GetTemplate", mock.Anything, "repo/my-chart", "0.1.0", "deployment").Return(result, nil)

	appContext := AppContext{}
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.hasConfigMap)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "GetTemplate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, "{\"result\":\"true\"}", string(rr.Body.Bytes()), "Response is not correct.")
}

func TestGetChartVariables(t *testing.T) {
	req, err := http.NewRequest("POST", "/getChartVariables", getPayloadChartRequest())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}

	partialResult := "{\"app\": {\"dateHour\": 0,\"pullSecret\": \"foo\"} }"
	mockHelmSvc.On("GetTemplate", mock.Anything, "foo", "0.1.0", "values").Return([]byte(partialResult), nil)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(getCharts())

	appContext := AppContext{}
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getChartVariables)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "GetTemplate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
	assert.Equal(t, partialResult, string(rr.Body.Bytes()), "Response is not correct.")
}

func TestGetChartVariables_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/getChartVariables", appContext.getChartVariables)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetChartVariables_GetTemplateError(t *testing.T) {
	req, err := http.NewRequest("POST", "/getChartVariables", getPayloadChartRequest())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, "foo", "0.1.0", "values").Return(nil, errors.New("some error"))
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(getCharts())

	appContext := AppContext{}
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getChartVariables)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "GetTemplate", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetHelmCommand(t *testing.T) {
	req, err := http.NewRequest("POST", "/getHelmCommand", getMultipleInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc := &mocks.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)
	appContext.HelmServiceAPI = mockHelmSvc

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getHelmCommand)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, "helm upgrade --install my-chart-dev")
	assert.Contains(t, response, "--set \"app.username=user")
	assert.Contains(t, response, "istio.virtualservices.gateways[0]=my-gateway.istio-system.svc.cluster.local")
	assert.Contains(t, response, "repo/my-chart - 0.1.0 --namespace=dev")
}

func TestGetHelmCommand_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadErrorWithPrincipal(t, "/getHelmCommand", appContext.getHelmCommand, "tenkai-helm-upgrade")
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetHelmCommand_GetByIDError(t *testing.T) {
	req, err := http.NewRequest("POST", "/getHelmCommand", getMultipleInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByIDError(&appContext)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getHelmCommand)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestMultipleInstall(t *testing.T) {
	multipleInstallPayload := getMultipleInstallPayload()
	req, err := http.NewRequest("POST", "/multipleInstall", multipleInstallPayload)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	charts := getCharts()

	appContext := AppContext{}

	mockDeploymentDAO := &mockRepo.DeploymentDAOInterface{}
	mockDeploymentDAO.On("CreateDeployment", mock.Anything).Return(1, nil)

	mockConfigDAO := &mockRepo.ConfigDAOInterface{}

	var config model.ConfigMap
	config.Name = "mykey"
	config.Value = "myvalue"
	mockConfigDAO.On("GetConfigByName", "commonValuesConfigMapChart").
		Return(config, nil)

	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockUpgrade(&appContext)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(charts)

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)

	user := mockUser()
	mockUserDAO := &mockRepo.UserDAOInterface{}
	mockUserDAO.On("FindByEmail", user.Email).Return(user, nil)

	secOper := mockSecurityOperations()
	mockUserEnvRoleDAO := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDAO.On("GetRoleByUserAndEnvironment", user, mock.Anything).
		Return(&secOper, nil)

	var p model.ProductVersion
	p.ID = uint(777)
	p.Version = "19.0.1-0"
	p.ProductID = 999
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductVersionsByID", mock.Anything).Return(&p, nil)

	pvs := make([]model.ProductVersionService, 0)
	pvs = append(pvs, getProductVersionSvcParams(999, "repo/my-chart - 0.1.0", 777, p.Version))
	mockProductDAO.On("ListProductsVersionServices", int(p.ID)).Return(pvs, nil)

	var varImgTag model.Variable
	varImgTag.Scope = "repo/my-chart"
	varImgTag.Name = "image.tag"
	varImgTag.Value = "18.0.1-0"
	varImgTag.EnvironmentID = 999

	mockVariableDAO.On("GetVarImageTagByEnvAndScope", 999, "repo/my-chart - 0.1.0").
		Return(varImgTag, nil)

	mockVariableDAO.On("EditVariable", mock.Anything).Return(nil)

	appContext.Repositories.UserDAO = mockUserDAO
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDAO
	appContext.Repositories.ConfigDAO = mockConfigDAO
	appContext.Repositories.DeploymentDAO = mockDeploymentDAO
	appContext.HelmServiceAPI = mockHelmSvc

	auditValues := make(map[string]string)
	auditValues["environment"] = "bar"
	auditValues["chartName"] = "repo/my-chart - 0.1.0"
	auditValues["name"] = "my-chart"
	mockAudit := mockDoAudit(&appContext, "deploy", auditValues)

	mockEnvDao.On("EditEnvironment", mock.Anything).Return(nil)

	webHooks := make([]model.WebHook, 0)
	webHooks = append(webHooks, mockWebHook())
	mockWebHookDAO := &mockRepo.WebHookDAOInterface{}
	mockWebHookDAO.On("ListWebHooksByEnvAndType", 999, "HOOK_DEPLOY_PRODUCT").
		Return(webHooks, nil)
	appContext.Repositories.WebHookDAO = mockWebHookDAO

	var product model.Product
	product.ID = 999
	product.Name = "My Product"
	mockProductDAO.On("FindProductByID", 999).Return(product, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.multipleInstall)

	appContext.RabbitImpl = getMockRabbitMQ()

	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 0)
	mockAudit.AssertNumberOfCalls(t, "DoAudit", 1)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetVarImageTagByEnvAndScope", 1)
	mockWebHookDAO.AssertNumberOfCalls(t, "ListWebHooksByEnvAndType", 1)
	mockProductDAO.AssertNumberOfCalls(t, "FindProductByID", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func getMockRabbitMQ() *mockRabbit.RabbitInterface {
	mockRabbitMQ := mockRabbit.RabbitInterface{}

	mockRabbitMQ.Mock.On("Publish",
		"",
		mock.Anything,
		false,
		false,
		mock.Anything,
	).Return(nil)

	return &mockRabbitMQ
}

func TestInstall(t *testing.T) {
	req, err := http.NewRequest("POST", "/install", getInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	charts := getCharts()

	appContext := AppContext{}

	mockDeploymentDAO := &mockRepo.DeploymentDAOInterface{}
	mockDeploymentDAO.On("CreateDeployment", mock.Anything).Return(1, nil)

	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockUpgrade(&appContext)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(charts)

	user := mockUser()
	mockUserDAO := &mockRepo.UserDAOInterface{}
	mockUserDAO.On("FindByEmail", user.Email).Return(user, nil)

	secOper := mockSecurityOperations()
	mockUserEnvRoleDAO := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDAO.On("GetRoleByUserAndEnvironment", user, mock.Anything).
		Return(&secOper, nil)

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)

	appContext.Repositories.UserDAO = mockUserDAO
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDAO
	appContext.Repositories.DeploymentDAO = mockDeploymentDAO

	appContext.RabbitImpl = getMockRabbitMQ()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.install)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 0)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestDryRun(t *testing.T) {
	req, err := http.NewRequest("POST", "/helmDryRun", getInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockUpgrade(&appContext)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(getCharts())

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)

	appContext.RabbitImpl = getMockRabbitMQ()

	mockDeploymentDAO := &mockRepo.DeploymentDAOInterface{}
	mockDeploymentDAO.On("CreateDeployment", mock.Anything).Return(1, nil)
	appContext.Repositories.DeploymentDAO = mockDeploymentDAO

	user := mockUser()
	mockUserDAO := &mockRepo.UserDAOInterface{}
	mockUserDAO.On("FindByEmail", mock.Anything).Return(user, nil)
	appContext.Repositories.UserDAO = mockUserDAO

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.helmDryRun)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 0)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func getInstallPayload() *bytes.Buffer {
	var payload model.InstallPayload
	payload.EnvironmentID = 999
	payload.Chart = "foo"
	payload.ChartVersion = "0.1.0"
	payload.Name = "my-foo"

	pStr, _ := json.Marshal(payload)
	return bytes.NewBuffer(pStr)
}

func mockUpgrade(appContext *AppContext) *mockSvc.HelmServiceInterface {
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("Upgrade", mock.Anything, mock.Anything).Return(nil)
	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func mockHelmSvcWithLotOfThings(appContext *AppContext) *mockSvc.HelmServiceInterface {

	hlr := helmapi.HelmListResult{}
	hlr.Releases = make([]helmapi.ListRelease, 0)
	lr := helmapi.ListRelease{}
	lr.Name = "tjusuarios-master"
	lr.Chart = "tjusuarios-master"
	lr.Namespace = "master"
	lr.Status = "Running"
	lr.AppVersion = "1.0"
	lr.Revision = 1
	hlr.Releases = append(hlr.Releases, lr)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("Upgrade", mock.Anything, mock.Anything).Return(nil)
	mockHelmSvc.On("ListHelmDeployments", mock.Anything, mock.Anything).Return(&hlr, nil)
	mockHelmSvc.On("DeleteHelmRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	repos := make([]model.Repository, 0)
	repo := model.Repository{}
	repo.Name = "alfa.beta"
	repo.Password = "123"
	repo.Username = "guri"
	repo.URL = "http://artifactory.xpto"
	repos = append(repos, repo)

	mockHelmSvc.On("GetRepositories").Return(repos, nil)
	mockHelmSvc.On("AddRepository", mock.Anything).Return(nil)

	services := make([]model.Service, 0)
	service := model.Service{}
	service.Name = "abacaxi"
	service.Age = "1d"
	service.Type = "ClusterIP"
	service.Ports = "1223, 8080"
	service.ExternalIP = ""
	services = append(services, service)

	pods := make([]model.Pod, 0)
	pod := model.Pod{}
	pod.Name = "alfa"
	pod.Age = "1d"
	pod.Status = "Running"
	pod.Restarts = 0
	pod.Image = "alfa"
	pod.Ready = "1/1"

	pods = append(pods, pod)

	mockHelmSvc.On("GetServices", mock.Anything, mock.Anything).Return(services, nil)
	mockHelmSvc.On("GetPods", mock.Anything, mock.Anything).Return(pods, nil)
	mockHelmSvc.On("DeletePod", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func getDeployable() []model.InstallPayload {
	var ip model.InstallPayload
	ip.EnvironmentID = 999
	ip.Chart = "repo/my-chart - 0.1.0"
	ip.ChartVersion = "0.1.0"
	ip.Name = "my-chart"
	return []model.InstallPayload{ip}
}

func getMultipleInstallPayload() *bytes.Buffer {
	var ip model.InstallPayload
	ip.EnvironmentID = 999
	ip.Chart = "repo/my-chart - 0.1.0"
	ip.ChartVersion = "0.1.0"
	ip.Name = "my-chart"

	var payload model.MultipleInstallPayload
	payload.EnvironmentIDs = []int{999}
	payload.ProductVersionID = 777
	payload.Deployables = append(payload.Deployables, ip)
	pStr, _ := json.Marshal(payload)
	return bytes.NewBuffer(pStr)
}

func getPayloadChartRequest() *bytes.Buffer {
	var p model.GetChartRequest
	p.ChartName = "repo/my-chart"
	p.ChartVersion = "0.1.0"
	pStr, _ := json.Marshal(p)
	return bytes.NewBuffer(pStr)
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
