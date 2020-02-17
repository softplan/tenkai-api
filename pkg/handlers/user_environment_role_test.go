package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserPolicyByEnvironment(t *testing.T) {
	appContext := AppContext{}

	var p model.GetUserPolicyByEnvironmentRequest
	p.EnvironmentID = 999
	p.Email = "test@example.com"

	var user model.User
	user.ID = 999
	user.Email = "test@example.com"

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	var policies []string
	policies = append(policies, "ACTION_DEPLOY")
	policies = append(policies, "ACTION_SAVE_VARIABLES")

	var result model.SecurityOperation
	result.ID = 888
	result.Name = "MASTER_OF_PUPPETS"
	result.Policies = policies

	mockUserEnvRoleDao := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDao.On("GetRoleByUserAndEnvironment", user, uint(p.EnvironmentID)).
		Return(&result, nil)

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDao

	req, err := http.NewRequest("POST", "/getUserPolicyByEnvironment", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getUserPolicyByEnvironment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")

	response := string(rr.Body.Bytes())
	fmt.Println(response)
	assert.Contains(t, response, `{"ID":888,`)
	assert.Contains(t, response, `"name":"MASTER_OF_PUPPETS",`)
	assert.Contains(t, response, `"policies":["ACTION_DEPLOY",`)
	assert.Contains(t, response, `"ACTION_SAVE_VARIABLES"]}`)
}

func TestGetUserPolicyByEnvironment_UnmarshalError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/getUserPolicyByEnvironment", appContext.getUserPolicyByEnvironment)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetUserPolicyByEnvironment_UserError(t *testing.T) {
	appContext := AppContext{}

	var p model.GetUserPolicyByEnvironmentRequest
	p.EnvironmentID = 999
	p.Email = "test@example.com"

	var user model.User
	user.ID = 999
	user.Email = "test@example.com"

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, errors.New("some error"))

	appContext.Repositories.UserDAO = mockUserDao

	req, err := http.NewRequest("POST", "/getUserPolicyByEnvironment", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getUserPolicyByEnvironment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestGetUserPolicyByEnvironment_RoleError(t *testing.T) {
	appContext := AppContext{}

	var p model.GetUserPolicyByEnvironmentRequest
	p.EnvironmentID = 999
	p.Email = "test@example.com"

	var user model.User
	user.ID = 999
	user.Email = "test@example.com"

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	var result model.SecurityOperation
	result.ID = 888

	mockUserEnvRoleDao := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDao.On("GetRoleByUserAndEnvironment", user, uint(p.EnvironmentID)).
		Return(&result, errors.New("some error"))

	appContext.Repositories.UserDAO = mockUserDao
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDao

	req, err := http.NewRequest("POST", "/getUserPolicyByEnvironment", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getUserPolicyByEnvironment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestCreateOrUpdateUserEnvironmentRole(t *testing.T) {
	appContext := AppContext{}
	var p model.UserEnvironmentRole
	p.UserID = 999
	p.EnvironmentID = 888
	p.SecurityOperationID = 777

	mockUserEnvRoleDao := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDao.On("CreateOrUpdate", p).Return(nil)
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDao

	req, err := http.NewRequest("POST", "/createOrUpdateUserEnvironmentRole", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateUserEnvironmentRole)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestCreateOrUpdateUserEnvironmentRole_UnmarshalError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/createOrUpdateUserEnvironmentRole", appContext.getUserPolicyByEnvironment)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCreateOrUpdateUserEnvironmentRoleError(t *testing.T) {
	appContext := AppContext{}
	var p model.UserEnvironmentRole
	p.UserID = 999
	p.EnvironmentID = 888
	p.SecurityOperationID = 777

	mockUserEnvRoleDao := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDao.On("CreateOrUpdate", p).Return(errors.New("some error"))
	appContext.Repositories.UserEnvironmentRoleDAO = mockUserEnvRoleDao

	req, err := http.NewRequest("POST", "/createOrUpdateUserEnvironmentRole", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateUserEnvironmentRole)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}
