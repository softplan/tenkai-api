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

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, MockGetEnv())
	mockEnvDao := MockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	auditValues := make(map[string]string)
	auditValues["variable_name"] = variable.Name
	auditValues["variable_old_value"] = ""
	auditValues["variable_new_value"] = variable.Value
	auditValues["scope"] = variable.Scope

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(auditValues, true, nil)

	mockAudit := MockDoAudit(&appContext, "saveVariable", auditValues)

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

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"role-unauthorized"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_GetByIDError(t *testing.T) {
	appContext := AppContext{}

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, MockGetEnv())
	mockEnvDao := MockGetByIDError(&appContext)

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)

	assert.Equal(t, http.StatusNotImplemented, rr.Code, "Response should be 501.")
}

func TestSaveVariableValues_HasAccessError(t *testing.T) {
	appContext := AppContext{}

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, MockGetEnv())
	mockEnvDao := MockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(nil, errors.New("Record not found"))

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestSaveVariableValues_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer([]byte(`["invalid": 123]`)))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSaveVariableValues_CreateVariableError(t *testing.T) {
	appContext := AppContext{}

	variable := MockVariable()
	var varData model.VariableData
	varData.Data = append(varData.Data, variable)

	payloadStr, _ := json.Marshal(varData)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, MockGetEnv())
	mockEnvDao := MockGetByID(&appContext)
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
