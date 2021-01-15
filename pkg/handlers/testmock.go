package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockAud "github.com/softplan/tenkai-api/pkg/audit/mocks"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/softplan/tenkai-api/pkg/service/core/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//mockPrincipal injects a http header with the specified role to be used only for testing.
func mockPrincipal(req *http.Request) {
	var roles []string
	roles = append(roles, "tenkai-admin")
	principal := model.Principal{Name: "alfa", Email: "beta@alfa.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))
}

//mockGetByID mocks a call to GetByID function to be used only for testing.
func mockGetByID(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := mockGetEnv()
	mockEnvDao.On("GetByID", int(env.ID)).Return(&env, nil)
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

func mockEnvDaoWithLotOfThings(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := mockGetEnv()
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)
	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(envs, nil)
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

//mockGetByIDError mocks a call to GetByID function returning an error to be used only for testing.
func mockGetByIDError(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	mockEnvDao.On("GetByID", mock.Anything).Return(nil, errors.New("Some error"))
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

//mockGetEnv returns an Environment struct to be used only for testing.
func mockGetEnv() model.Environment {
	var env model.Environment
	env.ID = 999
	env.Group = "foo"
	env.Name = "bar"
	env.ClusterURI = "https://rancher-k8s-my-domain.com/k8s/clusters/c-kbfxr"
	env.CACertificate = "my-certificate"
	env.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"
	env.Namespace = "dev"
	env.Gateway = "my-gateway.istio-system.svc.cluster.local"
	return env
}

//mockGlobalVariable returns a global Variable struct to be used only for testing.
func mockGlobalVariable() model.Variable {
	var variable model.Variable
	variable.Scope = "global"
	variable.Name = "username"
	variable.Value = "user"
	variable.Secret = false
	variable.Description = "Login username."
	variable.EnvironmentID = 999
	return variable
}

//mockVariable returns a Variable struct to be used only for testing.
func mockVariable() model.Variable {
	var variable model.Variable
	variable.Scope = "bar"
	variable.Name = "password"
	variable.Value = "password"
	variable.Secret = false
	variable.Description = "Login password."
	variable.EnvironmentID = 999
	return variable
}

//mockDoAudit mocks a call to DoAudit function to be used only for testing.
func mockDoAudit(appContext *AppContext, operation string, auditValues map[string]string) *mockAud.AuditingInterface {
	mockAudit := &mockAud.AuditingInterface{}
	mockAudit.On("DoAudit", mock.Anything, mock.Anything, "beta@alfa.com", operation, auditValues)
	appContext.Auditing = mockAudit

	return mockAudit
}

func mockGetAllVariablesByEnvironmentAndScope(appContext *AppContext) *mockRepo.VariableDAOInterface {
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	var variables []model.Variable
	variable := mockGlobalVariable()
	variables = append(variables, variable)
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", int(variable.EnvironmentID), mock.Anything).Return(variables, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
}

func mockVariableDAOWithLotOfThings(appContext *AppContext) *mockRepo.VariableDAOInterface {
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	var variables []model.Variable
	variable := mockGlobalVariable()
	variables = append(variables, variable)
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", int(variable.EnvironmentID), mock.Anything).Return(variables, nil)
	mockVariableDAO.On("DeleteVariableByEnvironmentID", mock.Anything).Return(nil)
	mockVariableDAO.On("GetAllVariablesByEnvironment", mock.Anything).Return(variables, nil)
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, true, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
}

func mockGetAllVariablesByEnvironment(appContext *AppContext) *mockRepo.VariableDAOInterface {
	var variables []model.Variable
	variables = append(variables, mockGlobalVariable())
	variables = append(variables, mockVariable()) // Not used variable
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("GetAllVariablesByEnvironment", mock.Anything).Return(variables, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
}

func mockGetAllVariablesByEnvironmentError(appContext *AppContext) *mockRepo.VariableDAOInterface {
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("GetAllVariablesByEnvironment", mock.Anything).Return(nil, errors.New("some error"))
	appContext.Repositories.VariableDAO = mockVariableDAO
	return mockVariableDAO
}

func mockGetAllVariablesByEnvironmentAndScopeError(appContext *AppContext) *mockRepo.VariableDAOInterface {
	variable := mockGlobalVariable()
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", int(variable.EnvironmentID), mock.Anything).Return(nil, errors.New("Some error"))

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
}

//testHandlerFunc should be used only for testing.
type testHandlerFunc func(http.ResponseWriter, *http.Request)

func testUnmarshalPayloadError(t *testing.T, endpoint string, handFunc testHandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`["invalid": 123]`)))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handFunc)
	handler.ServeHTTP(rr, req)

	return rr
}

func testUnmarshalPayloadErrorWithPrincipal(t *testing.T, endpoint string, handFunc testHandlerFunc, role string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`["invalid": 123]`)))
	assert.NoError(t, err)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handFunc)
	handler.ServeHTTP(rr, req)

	return rr
}

func commonTestHasAccessError(t *testing.T, endpoint string, handFunc testHandlerFunc, appContext *AppContext) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`{"data":[{"environmentId":999}]}`)))
	assert.NoError(t, err)

	mockPrincipal(req)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(nil, errors.New("Record not found"))

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handFunc)
	handler.ServeHTTP(rr, req)

	return rr
}

func mockGetAllEnvironments(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	return mockGetAllEnvironmentsPrincipal(appContext, "beta@alfa.com")
}

func mockGetAllEnvironmentsPrincipal(appContext *AppContext, principal string) *mockRepo.EnvironmentDAOInterface {
	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	mockEnvDao.On("GetAllEnvironments", principal).Return(envs, nil)
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

func mockGetAllEnvironmentsError(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(nil, errors.New("some error"))
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

func mockConventionInterface(appContext *AppContext) *mocks.ConventionInterface {
	mockConvention := &mocks.ConventionInterface{}
	mockConvention.On("GetKubeConfigFileName", "foo", "bar").Return("./config/foo_bar")
	appContext.ConventionInterface = mockConvention
	return mockConvention
}

func mockHelmListResult() *helmapi.HelmListResult {
	var listReleases []helmapi.ListRelease

	lr := helmapi.ListRelease{
		Name:       "my-foo",
		Revision:   9999,
		Updated:    "",
		Status:     "",
		Chart:      "foo",
		AppVersion: "0.1.0",
		Namespace:  "dev",
	}
	listReleases = append(listReleases, lr)

	result := &helmapi.HelmListResult{
		Next:     "998",
		Releases: listReleases,
	}
	return result
}

func mockListHelmDeployments(appContext *AppContext) *mockSvc.HelmServiceInterface {
	result := mockHelmListResult()

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("ListHelmDeployments", mock.Anything, "dev").Return(result, nil)

	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func mockListHelmDeploymentsError(appContext *AppContext) *mockSvc.HelmServiceInterface {
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("ListHelmDeployments", mock.Anything, "dev").Return(nil, errors.New("some error"))

	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func payload(v interface{}) *bytes.Buffer {
	payloadStr, _ := json.Marshal(v)
	return bytes.NewBuffer(payloadStr)
}

func getProduct() model.Product {
	var payload model.Product
	payload.ID = 999
	payload.Name = "my-product"
	payload.ValidateReleases = true
	return payload
}

func getProductWithoutID() model.Product {
	var payload model.Product
	payload.Name = "my-product"
	return payload
}

func getProductVersionWithoutID(baseRelease int) model.ProductVersion {
	var p model.ProductVersion
	p.Version = "19.0.1-0"
	p.ProductID = 999
	p.BaseRelease = baseRelease
	p.Locked = false
	return p
}

func getProductVersionSvc() model.ProductVersionService {
	return getProductVersionSvcParams(888, "repo/my-chart - 0.1.0", 999, "19.0.1-0")
}

func getProductVersionSvcParams(id int, svcName string, productID int, tag string) model.ProductVersionService {
	var pvs model.ProductVersionService
	pvs.ID = 888
	pvs.ServiceName = "repo/my-chart - 0.1.0"
	pvs.ProductVersionID = 999
	pvs.DockerImageTag = "19.0.1-0"
	return pvs
}

func getProductVersionSvcReqResp() *model.ProductVersionServiceRequestReponse {
	childs := &model.ProductVersionServiceRequestReponse{}
	childs.List = append(childs.List, getProductVersionSvc())
	return childs
}

func getProductVersionReqResp() *model.ProductVersionRequestReponse {
	pv := getProductVersionWithoutID(0)
	pv.ID = 777
	l := &model.ProductVersionRequestReponse{}
	l.List = append(l.List, pv)
	return l
}

func getHelmSearchResult() []model.SearchResult {
	data := make([]model.SearchResult, 1)
	data[0].Name = "repo/my-chart"
	data[0].ChartVersion = "1.0"
	data[0].Description = "Test only"
	data[0].AppVersion = "1.0"
	return data
}

func mockHelmSearchCharts(appContext *AppContext) *mockSvc.HelmServiceInterface {
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	data := getHelmSearchResult()
	mockHelmSvc.On("SearchCharts", mock.Anything, true).Return(&data)
	appContext.HelmServiceAPI = mockHelmSvc
	return mockHelmSvc
}

func mockEditVariableError(appContext *AppContext) *mockRepo.VariableDAOInterface {
	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("EditVariable", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.VariableDAO = mockVariableDAO
	return mockVariableDAO
}

func mockVariableRule() model.VariableRule {
	var item model.VariableRule
	item.Name = "urlapi.*"
	return item
}

func mockVariableRuleWithID() model.VariableRule {
	vr := mockValueRuleWithID()

	var item model.VariableRule
	item.ID = 999
	item.Name = "urlapi.*"
	item.ValueRules = append(item.ValueRules, &vr)
	return item
}

func mockValueRule() model.ValueRule {
	var vr model.ValueRule
	vr.Value = "http"
	vr.Type = "StartsWith"
	vr.VariableRuleID = 999
	return vr
}

func mockValueRuleWithID() model.ValueRule {
	var vr model.ValueRule
	vr.ID = 888
	vr.Value = "http"
	vr.Type = "StartsWith"
	vr.VariableRuleID = 999
	return vr
}

func getVarByEnvAndScopePayload() *bytes.Buffer {
	return createPayloadWithScopeAndID(999, "global")
}

func createPayloadWithScopeAndID(id int, scope string) *bytes.Buffer {
	type Payload struct {
		EnvironmentID int    `json:"environmentId"`
		Scope         string `json:"scope"`
	}

	var payload Payload
	payload.EnvironmentID = id
	payload.Scope = scope
	payloadStr, _ := json.Marshal(payload)

	return bytes.NewBuffer(payloadStr)
}

func mockPolicies() []string {
	var policies []string
	policies = append(policies, "ACTION_DEPLOY")
	return policies
}

func mockSecurityOperations() model.SecurityOperation {
	var so model.SecurityOperation
	so.ID = 999
	so.Name = "ONLY_DEPLOY"
	so.Policies = mockPolicies()
	return so
}

func mockUser() model.User {
	var user model.User
	user.ID = 999
	user.Email = "beta@alfa.com"
	return user
}

func getUserPolicyByEnv() model.GetUserPolicyByEnvironmentRequest {
	var p model.GetUserPolicyByEnvironmentRequest
	p.EnvironmentID = 999
	p.Email = "beta@alfa.com"
	return p
}

func mockUserEnvRole() model.UserEnvironmentRole {
	var u model.UserEnvironmentRole
	u.UserID = 999
	u.EnvironmentID = 888
	u.SecurityOperationID = 777
	return u
}

func mockWebHook() model.WebHook {
	var item model.WebHook
	item.Name = "Product Deploy"
	item.Type = "HOOK_DEPLOY_PRODUCT"
	item.URL = "http://localhost/incoming-hook"
	item.EnvironmentID = 999
	return item
}
