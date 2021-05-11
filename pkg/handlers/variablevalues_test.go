package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveVariableValues(t *testing.T) {
	appContext := AppContext{}

	variable := mockGlobalVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc := &mocks.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)
	appContext.HelmServiceAPI = mockHelmSvc

	auditValues := make(map[string]string)
	auditValues["variable_name"] = variable.Name
	auditValues["variable_old_value"] = ""
	auditValues["variable_new_value"] = variable.Value
	auditValues["scope"] = variable.Scope

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("CreateVariableWithDefaultValue", mock.Anything).Return(auditValues, true, nil)
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(auditValues, true, nil)

	mockAudit := mockDoAudit(&appContext, "saveVariable", auditValues)

	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.Repositories.VariableDAO = mockVariableDAO

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "CreateVariable", 1)
	mockAudit.AssertNumberOfCalls(t, "DoAudit", 2)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created.")
}

func TestSaveVariableValues_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	user := mockUser()
	mockUserDAO := &mockRepo.UserDAOInterface{}
	mockUserDAO.On("FindByEmail", mock.Anything).Return(user, nil)

	secOper := mockSecurityOperations()
	secOper.Policies = make([]string, 0)
	mockUserEnvRoleDAO := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDAO.On("GetRoleByUserAndEnvironment", user, mock.Anything).
		Return(&secOper, nil)

	mockGetByID(&appContext)

	appContext.Repositories.UserDAO = mockUserDAO
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDAO

	variable := mockGlobalVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_GetByIDError(t *testing.T) {
	appContext := AppContext{}

	variable := mockGlobalVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByIDError(&appContext)

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Response should be 400.")
}

func TestSaveVariableValues_HasAccessError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestHasAccessError(t, "/saveVariableValues", appContext.saveVariableValues, &appContext)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadErrorWithPrincipal(t, "/saveVariableValues", appContext.saveVariableValues, "tenkai-variables-save")
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSaveVariableValues_CreateVariableError(t *testing.T) {
	appContext := AppContext{}

	variable := mockGlobalVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req)

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc := &mocks.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)
	appContext.HelmServiceAPI = mockHelmSvc

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, false, errors.New("Error saving variable"))
	mockVariableDAO.On("CreateVariableWithDefaultValue", mock.Anything).Return(nil, false, nil)

	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.Repositories.VariableDAO = mockVariableDAO

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "CreateVariable", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetVariablesByEnvironmentAndScope(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("POST", "/listVariables", getVarByEnvAndScopePayload())
	assert.NoError(t, err)

	mockPrincipal(req)

	mockEnvDao := mockGetAllEnvironments(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getVariablesByEnvironmentAndScope)
	handler.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"Variables":[{"ID":0,`)
	assert.Contains(t, response, `"scope":"global","chartVersion":"","name":"username","value":"user",`)
	assert.Contains(t, response, `"secret":false,"description":"Login username.","environmentId":999}]}`)
}

func TestGetVariablesByEnvironmentAndScope_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/listVariables", appContext.getVariablesByEnvironmentAndScope)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetVariablesByEnvironmentAndScope_HasAccessError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestHasAccessError(t, "/listVariables", appContext.getVariablesByEnvironmentAndScope, &appContext)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestGetVariablesByEnvironmentAndScope_Error(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("POST", "/listVariables", getVarByEnvAndScopePayload())
	assert.NoError(t, err)

	mockPrincipal(req)

	mockEnvDao := mockGetAllEnvironments(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScopeError(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getVariablesByEnvironmentAndScope)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetVariablesNotUsed(t *testing.T) {
	appContext := AppContext{}

	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)

	mockEnvDao := mockGetByID(&appContext)
	appContext.Repositories.VariableDAO = mockVariableDAO

	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockListHelmDeployments(&appContext)

	req, err := http.NewRequest("GET", "/getVariablesNotUsed/999", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/getVariablesNotUsed/{id}", appContext.getVariablesNotUsed).Methods("GET")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "ListHelmDeployments", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	assert.Equal(t, `[{"id":0,"scope":"bar","name":"password","value":"password"}]`,
		string(rr.Body.Bytes()), "Should found 1 not used variable.")
}
