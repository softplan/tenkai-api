package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
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

	mockObject.On("SearchCharts", mock.Anything, true).Return(&data)
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

	mockPrincipal(req, []string{"tenkai-helm-upgrade"})

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
	info.Chart = "my-_helm-chart"
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

func TestHasConfigMap(t *testing.T) {
	req, err := http.NewRequest("POST", "/hasConfigMap", getPayloadChartRequest())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}

	partialDeploymentYaml := "...name: {{ template \"foo.name\" . }}-gcm-{{ .Release.Namespace }}..."
	result := []byte(partialDeploymentYaml)
	mockHelmSvc.On("GetTemplate", mock.Anything, "foo", "0.1.0", "deployment").Return(result, nil)

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

	appContext := AppContext{}
	appContext.HelmServiceAPI = mockHelmSvc

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getChartVariables)
	handler.ServeHTTP(rr, req)

	mockHelmSvc.AssertNumberOfCalls(t, "GetTemplate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, partialResult, string(rr.Body.Bytes()), "Response is not correct.")
}

func TestGetHelmCommand(t *testing.T) {
	req, err := http.NewRequest("POST", "/getHelmCommand", getMultipleInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)

	mockPrincipal(req, []string{"tenkai-helm-upgrade"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getHelmCommand)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, "_helm upgrade --install my-foo-dev")
	assert.Contains(t, response, "--set \"app.username=user")
	assert.Contains(t, response, "istio.virtualservices.gateways[0]=my-gateway.istio-system.svc.cluster.local")
	assert.Contains(t, response, "foo --namespace=dev")
}

func TestMultipleInstall(t *testing.T) {
	req, err := http.NewRequest("POST", "/multipleInstall", getMultipleInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-helm-upgrade"})

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockUpgrade(&appContext)

	auditValues := make(map[string]string)
	auditValues["environment"] = "bar"
	auditValues["chartName"] = "foo"
	auditValues["name"] = "my-foo"
	mockAudit := mockDoAudit(&appContext, "deploy", auditValues)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.multipleInstall)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 1)
	mockAudit.AssertNumberOfCalls(t, "DoAudit", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestInstall(t *testing.T) {
	req, err := http.NewRequest("POST", "/install", getInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-helm-upgrade"})

	appContext := AppContext{}
	mockEnvDao := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockUpgrade(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.install)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 1)

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

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.helmDryRun)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)
	mockHelmSvc.AssertNumberOfCalls(t, "Upgrade", 1)

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

func getMultipleInstallPayload() *bytes.Buffer {
	var ip model.InstallPayload
	ip.EnvironmentID = 999
	ip.Chart = "foo"
	ip.ChartVersion = "0.1.0"
	ip.Name = "my-foo"

	var payload model.MultipleInstallPayload
	payload.Deployables = append(payload.Deployables, ip)
	pStr, _ := json.Marshal(payload)
	return bytes.NewBuffer(pStr)
}

func getPayloadChartRequest() *bytes.Buffer {
	var p model.GetChartRequest
	p.ChartName = "foo"
	p.ChartVersion = "0.1.0"
	pStr, _ := json.Marshal(p)
	return bytes.NewBuffer(pStr)
}

func getExpect(sr []model.SearchResult) string {
	j, _ := json.Marshal(sr)
	return "{\"charts\":" + string(j) + "}"
}

func getExpecHistory() string {
	return "[{\"revision\":987,\"updated\":\"\",\"status\":\"DEPLOYED\",\"chart\":\"my-_helm-chart\",\"description\":\"Install completed\"}]"
}

func getRevision() *model.GetRevisionRequest {
	var revision model.GetRevisionRequest
	revision.EnvironmentID = 999
	revision.ReleaseName = "foo"
	revision.Revision = 800
	return &revision
}
