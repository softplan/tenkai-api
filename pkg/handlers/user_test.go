package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewUser(t *testing.T) {
	var p model.User
	p.Email = "alfa@beta.com.br"
	p.DefaultEnvironmentID = 1
	p.Environments = make([]model.Environment, 0)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateUser", mock.Anything).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/user", payload(p))
	assert.NoError(t, err)

	mockPrincipal(req, constraints.TenkaiAdmin)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newUser)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "CreateUser", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be 201.")
}

func TestNewUser_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("POST", "/user", nil)
	assert.NoError(t, err)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestNewUser_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadErrorWithPrincipal(t, "/user", appContext.newUser, "tenkai-admin")
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewUser_Error(t *testing.T) {
	var p model.User
	p.Email = "alfa@beta.com.br"
	p.DefaultEnvironmentID = 1
	p.Environments = make([]model.Environment, 0)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateUser", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/user", payload(p))
	assert.NoError(t, err)

	mockPrincipal(req, constraints.TenkaiAdmin)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newUser)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "CreateUser", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCreateOrUpdateUser(t *testing.T) {

	var p model.User
	p.Email = "alfa@beta.com.br"
	p.DefaultEnvironmentID = 1
	p.Environments = make([]model.Environment, 0)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateOrUpdateUser", mock.Anything).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/users/createOrUpdate", payload(p))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "CreateOrUpdateUser", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be 201.")
}

func TestCreateOrUpdateUser_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/users/createOrUpdate", appContext.createOrUpdateUser)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCreateOrUpdateUser_Error(t *testing.T) {

	var p model.User
	p.Email = "alfa@beta.com.br"
	p.DefaultEnvironmentID = 1
	p.Environments = make([]model.Environment, 0)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateOrUpdateUser", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/users/createOrUpdate", payload(p))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "CreateOrUpdateUser", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListUsers(t *testing.T) {
	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	result := &model.UserResult{}

	userDAO.On("ListAllUsers").Return(result.Users, nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("GET", "/users", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listUsers)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "ListAllUsers", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok.")
}

func TestListUsers_Error(t *testing.T) {
	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}

	userDAO.On("ListAllUsers").Return(nil, errors.New("some error"))
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("GET", "/users", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listUsers)
	handler.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "ListAllUsers", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteUser(t *testing.T) {

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}

	userDAO.On("DeleteUser", mock.AnythingOfType("int")).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("DELETE", "/users/9999", nil)
	assert.NoError(t, err)

	mockPrincipal(req, constraints.TenkaiAdmin)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")
	r.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "DeleteUser", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok.")
}

func TestDeleteUser_Unauthorized(t *testing.T) {

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}

	userDAO.On("DeleteUser", mock.AnythingOfType("int")).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("DELETE", "/users/9999", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")
	r.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "DeleteUser", 1)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func TestDeleteUser_Error(t *testing.T) {

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}

	userDAO.On("DeleteUser", mock.AnythingOfType("int")).Return(errors.New("some error"))
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("DELETE", "/users/9999", nil)
	assert.NoError(t, err)

	mockPrincipal(req, constraints.TenkaiAdmin)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")
	r.ServeHTTP(rr, req)

	userDAO.AssertNumberOfCalls(t, "DeleteUser", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}
