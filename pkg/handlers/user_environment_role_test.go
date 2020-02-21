package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserPolicyByEnvironment(t *testing.T) {
	appContext := AppContext{}

	p := getUserPolicyByEnv()

	user := mockUser()
	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	result := mockSecurityOperations()
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
	assert.Contains(t, response, `{"ID":999,`)
	assert.Contains(t, response, `"name":"ONLY_DEPLOY",`)
	assert.Contains(t, response, `"policies":["ACTION_DEPLOY"]}`)
}

func TestGetUserPolicyByEnvironment_UnmarshalError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/getUserPolicyByEnvironment", appContext.getUserPolicyByEnvironment)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetUserPolicyByEnvironment_UserError(t *testing.T) {
	appContext := AppContext{}

	p := getUserPolicyByEnv()
	user := mockUser()

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

	p := getUserPolicyByEnv()
	user := mockUser()

	mockUserDao := &mockRepo.UserDAOInterface{}
	mockUserDao.On("FindByEmail", mock.Anything).Return(user, nil)

	mockUserEnvRoleDao := &mockRepo.UserEnvironmentRoleDAOInterface{}
	mockUserEnvRoleDao.On("GetRoleByUserAndEnvironment", user, uint(p.EnvironmentID)).
		Return(nil, errors.New("some error"))

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

	p := mockUserEnvRole()

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
	rr := testUnmarshalPayloadError(t, "/createOrUpdateUserEnvironmentRole", appContext.createOrUpdateUserEnvironmentRole)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCreateOrUpdateUserEnvironmentRoleError(t *testing.T) {
	appContext := AppContext{}

	p := mockUserEnvRole()

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
