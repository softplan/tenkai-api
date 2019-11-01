package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSaveVariableValues(t *testing.T) {
	appContext := AppContext{}

	variable := mockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	auditValues := make(map[string]string)
	auditValues["variable_name"] = variable.Name
	auditValues["variable_old_value"] = ""
	auditValues["variable_new_value"] = variable.Value
	auditValues["scope"] = variable.Scope

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
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
	mockAudit.AssertNumberOfCalls(t, "DoAudit", 1)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created.")
}

func TestSaveVariableValues_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	variable := mockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"role-unauthorized"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_GetByIDError(t *testing.T) {
	appContext := AppContext{}

	variable := mockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByIDError(&appContext)

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusNotImplemented, rr.Code, "Response should be 501.")
}

func TestSaveVariableValues_HasAccessError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestHasAccessError(t, "/saveVariableValues", appContext.saveVariableValues, &appContext)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestUnmarshalPayloadError(t, "/saveVariableValues", appContext.saveVariableValues)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSaveVariableValues_CreateVariableError(t *testing.T) {
	appContext := AppContext{}

	variable := mockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, false, errors.New("Error saving variable"))

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

	type Payload struct {
		EnvironmentID int    `json:"environmentId"`
		Scope         string `json:"scope"`
	}

	var payload Payload
	payload.EnvironmentID = 999
	payload.Scope = "global"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/listVariables", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	mockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getVariablesByEnvironmentAndScope)
	handler.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironmentAndScope", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"Variables":[{"ID":0,`)
	assert.Contains(t, response, `"scope":"global","name":"username","value":"user",`)
	assert.Contains(t, response, `"secret":false,"description":"Login username.","environmentId":999}]}`)
}

func TestGetVariablesByEnvironmentAndScope_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestUnmarshalPayloadError(t, "/listVariables", appContext.getVariablesByEnvironmentAndScope)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetVariablesByEnvironmentAndScope_HasAccessError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestHasAccessError(t, "/listVariables", appContext.getVariablesByEnvironmentAndScope, &appContext)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}
