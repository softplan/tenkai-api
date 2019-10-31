package handlers

import (
	"bytes"
	"encoding/json"
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

	var vd model.VariableData
	vd.Data = append(vd.Data, MockVariable())

	payloadStr, _ := json.Marshal(vd)
	req, err := http.NewRequest("POST", "/saveVariableValues", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)

	MockPrincipal(req, []string{"tenkai-variables-save"})

	var envs []model.Environment
	envs = append(envs, MockGetEnv())
	mockEnvDao := MockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	mockVariableDAO := &mockRepo.VariableDAOInterface{}
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, true, nil)

	appContext.Repositories.EnvironmentDAO = mockEnvDao

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.saveVariableValues)
	handler.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not Created.")
}
