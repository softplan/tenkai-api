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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//mockPrincipal injects a http header with the specified role to be used only for testing.
func mockPrincipal(req *http.Request, roles []string) {
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
	env.Token = "my-token"
	env.Namespace = "dev"
	env.Gateway = "my-gateway.istio-system.svc.cluster.local"
	return env
}

//mockVariable returns a Variable struct to be used only for testing.
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
	variable := mockVariable()
	variables = append(variables, variable)
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", int(variable.EnvironmentID), mock.Anything).Return(variables, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	return mockVariableDAO
}

//testHandlerFunc should be used only for testing.
type testHandlerFunc func(http.ResponseWriter, *http.Request)

func commonTestUnmarshalPayloadError(t *testing.T, endpoint string, handFunc testHandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`["invalid": 123]`)))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handFunc)
	handler.ServeHTTP(rr, req)

	return rr
}

func commonTestHasAccessError(t *testing.T, endpoint string, handFunc testHandlerFunc, appContext *AppContext) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`{"data":[{"environmentId":999}]}`)))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

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
