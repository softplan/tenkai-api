package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEnvironmentPermission(t *testing.T) {

	appContext := AppContext{}

	mockUserDAO := mocks.UserDAOInterface{}
	mockUserDAO.On("AssociateEnvironmentUser", mock.Anything, mock.Anything).Return(nil)

	appContext.Repositories.UserDAO = &mockUserDAO

	req, err := http.NewRequest("GET", "/permissions/users/10/environments/99", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Status Created.")

	mockUserDAO.AssertNumberOfCalls(t, "AssociateEnvironmentUser", 1)

}

func TestNewEnvironmentPermission_UserIDError(t *testing.T) {

	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/permissions/users/QWERT/environments/99", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewEnvironmentPermission_EnvironmentIDError(t *testing.T) {

	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/permissions/users/10/environments/QWERT", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewEnvironmentPermission_Unauthorized(t *testing.T) {

	appContext := AppContext{}

	mockUserDAO := mocks.UserDAOInterface{}
	mockUserDAO.On("AssociateEnvironmentUser", mock.Anything, mock.Anything).Return(nil)

	appContext.Repositories.UserDAO = &mockUserDAO

	req, err := http.NewRequest("GET", "/permissions/users/10/environments/99", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestNewEnvironmentPermission_Error(t *testing.T) {

	appContext := AppContext{}

	mockUserDAO := mocks.UserDAOInterface{}
	mockUserDAO.On("AssociateEnvironmentUser", mock.Anything, mock.Anything).Return(errors.New("some error"))

	appContext.Repositories.UserDAO = &mockUserDAO

	req, err := http.NewRequest("GET", "/permissions/users/10/environments/99", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
	mockUserDAO.AssertNumberOfCalls(t, "AssociateEnvironmentUser", 1)

}
