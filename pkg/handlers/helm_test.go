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

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(envs, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("DeleteHelmRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	mockAudit.AssertExpectations(t)

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
	mockHelmSvc.On("RollbackRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
	mockHelmSvc.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(yaml, nil)

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
	info.Chart = "my-helm-chart"
	info.Description = "Install completed"

	var history helmapi.ReleaseHistory
	history = append(history, info)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetHelmReleaseHistory", mock.Anything, mock.Anything).Return(history, nil)

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
	req, err := http.NewRequest("POST", "/hasConfigMap", getPayloadChartRequest())
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

func TestGetChartVariables(t *testing.T) {
	req, err := http.NewRequest("POST", "/getChartVariables", getPayloadChartRequest())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}

	partialResult := "{\"app\": {\"dateHour\": 0,\"pullSecret\": \"foo\"} }"
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(partialResult), nil)

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

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getHelmCommand)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 2)

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, "helm upgrade --install my-foo-dev")
	assert.Contains(t, response, "--set \"app.username=user")
	assert.Contains(t, response, "istio.virtualservices.gateways[0]=my-gateway.istio-system.svc.cluster.local")
	assert.Contains(t, response, "foo --namespace=dev")
}

func TestMultipleInstall(t *testing.T) {
	req, err := http.NewRequest("POST", "/multipleInstall", getMultipleInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

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
	mockAudit.AssertExpectations(t)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestInstall(t *testing.T) {
	req, err := http.NewRequest("POST", "/install", getInstallPayload())
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

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

func getInstallPayload() *bytes.Buffer {
	var payload model.InstallPayload
	payload.EnvironmentID = 999
	payload.Chart = "foo"
	payload.ChartVersion = "0.1.0"
	payload.Name = "my-foo"

	pStr, _ := json.Marshal(payload)
	return bytes.NewBuffer(pStr)
}

func mockDoAudit(appContext *AppContext, operation string, auditValues map[string]string) *mockAud.AuditingInterface {
	mockAudit := &mockAud.AuditingInterface{}
	mockAudit.On("DoAudit", mock.Anything, mock.Anything, "beta@alfa.com", operation, auditValues)
	appContext.Auditing = mockAudit

	return mockAudit
}

func mockUpgrade(appContext *AppContext) *mockSvc.HelmServiceInterface {
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("Upgrade", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func mockConventionInterface(appContext *AppContext) *mocks.ConventionInterface {
	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", mock.Anything, mock.Anything).Return("./config/foo_bar")
	appContext.ConventionInterface = mockConvention
	return mockConvention
}

func mockGetAllVariablesByEnvironmentAndScope(appContext *AppContext) *mockRepo.VariableDAOInterface {
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	var variables []model.Variable
	variables = append(variables, mockVariable())
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", mock.Anything, mock.Anything).Return(variables, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
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

func mockVariable() model.Variable {
	var variable model.Variable
	variable.Scope = "global"
	variable.Name = "username"
	variable.Value = "user"
	variable.Secret = false
	variable.Description = "Login username."
	variable.EnvironmentID = 999
	return variable
}

func mockGetEnv() model.Environment {
	var env model.Environment
	env.ID = 999
	env.Group = "foo"
	env.Name = "bar"
	env.ClusterURI = "https://rancher-k8s-my-domain.com/k8s/clusters/c-kbfxr"
	env.CACertificate = "my-certificate"
	env.Token = "my-token"
	env.Namespace = "dev"
	env.Gateway = "my-gateway.istio-system.svc.cluster.local"
	return env
}

func mockGetByID(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := mockGetEnv()
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

func mockPrincipal(req *http.Request) {
	roles := []string{"tenkai-helm-upgrade"}
	principal := model.Principal{Name: "alfa", Email: "beta@alfa.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))
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
	return "[{\"revision\":987,\"updated\":\"\",\"status\":\"DEPLOYED\",\"chart\":\"my-helm-chart\",\"description\":\"Install completed\"}]"
}

func getRevision() *model.GetRevisionRequest {
	var revision model.GetRevisionRequest
	revision.EnvironmentID = 999
	revision.ReleaseName = "foo"
	revision.Revision = 800
	return &revision
}
