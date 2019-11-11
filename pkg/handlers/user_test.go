package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewUser(t *testing.T) {

	var payload model.User
	payload.Email = "alfa@beta.com.br"
	payload.DefaultEnvironmentID = 1
	payload.Environments = make([]model.Environment, 0)

	payS, _ := json.Marshal(payload)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateUser", mock.Anything).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/user", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newUser)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	userDAO.AssertNumberOfCalls(t, "CreateUser", 1)

}

func TestCreateOrUpdateUser(t *testing.T) {

	var payload model.User
	payload.Email = "alfa@beta.com.br"
	payload.DefaultEnvironmentID = 1
	payload.Environments = make([]model.Environment, 0)

	payS, _ := json.Marshal(payload)

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	userDAO.On("CreateOrUpdateUser", mock.Anything).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("POST", "/users/createOrUpdate", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateUser)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	userDAO.AssertNumberOfCalls(t, "CreateOrUpdateUser", 1)

}

func TestListUsers(t *testing.T) {

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}
	result := &model.UserResult{}

	userDAO.On("ListAllUsers").Return(result.Users, nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listUsers)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	userDAO.AssertNumberOfCalls(t, "ListAllUsers", 1)

}

func TestDeleteUser(t *testing.T) {

	appContext := AppContext{}

	userDAO := mocks.UserDAOInterface{}

	userDAO.On("DeleteUser", mock.AnythingOfType("int")).Return(nil)
	appContext.Repositories.UserDAO = &userDAO

	req, err := http.NewRequest("DELETE", "/users/9999", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")
	r.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	userDAO.AssertNumberOfCalls(t, "DeleteUser", 1)

}
