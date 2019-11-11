package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockDbms "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/softplan/tenkai-api/pkg/service/docker/mocks"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDockerTags(t *testing.T) {

	var payload model.ListDockerTagsRequest
	payload.ImageName = "alfa/beta"
	payload.From = "1.0"
	payS, _ := json.Marshal(payload)

	appContext := AppContext{}

	dockerMockSvc := mocks.DockerServiceInterface{}

	result := model.ListDockerTagsResult{}
	result.TagResponse = make([]model.TagResponse, 0)

	dockerMockSvc.On("GetDockerTagsWithDate", mock.Anything, mock.Anything, mock.Anything).Return(&result, nil)

	appContext.DockerServiceAPI = &dockerMockSvc

	req, err := http.NewRequest("POST", "/listDockerTags", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDockerTags)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	dockerMockSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)

}

func TestListDockerRepositories(t *testing.T) {

	appContext := AppContext{}

	dockerDAO := mockDbms.DockerDAOInterface{}

	repo := make([]model.DockerRepo, 0)

	dockerDAO.On("ListDockerRepos").Return(repo, nil)

	appContext.Repositories.DockerDAO = &dockerDAO

	result := model.ListDockerTagsResult{}
	result.TagResponse = make([]model.TagResponse, 0)

	req, err := http.NewRequest("GET", "/listDockerRepositories", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDockerRepositories)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	dockerDAO.AssertNumberOfCalls(t, "ListDockerRepos", 1)

}

func TestNewDockerRepository(t *testing.T) {

	var payload model.DockerRepo

	payload.Password = "123456"
	payload.Username = "alfa"
	payload.Host = "beta.com.br"
	payS, _ := json.Marshal(payload)

	appContext := AppContext{}

	dockerDAO := mockDbms.DockerDAOInterface{}
	dockerDAO.On("CreateDockerRepo", mock.Anything).Return(1, nil)

	appContext.Repositories.DockerDAO = &dockerDAO

	req, err := http.NewRequest("POST", "/newDockerRepository", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}
	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newDockerRepository)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	dockerDAO.AssertNumberOfCalls(t, "CreateDockerRepo", 1)

}

func TestDeleteDockerRepository(t *testing.T) {

	appContext := AppContext{}

	dockerDAO := mockDbms.DockerDAOInterface{}

	dockerDAO.On("DeleteDockerRepo", mock.AnythingOfType("int")).Return(nil)

	appContext.Repositories.DockerDAO = &dockerDAO

	result := model.ListDockerTagsResult{}
	result.TagResponse = make([]model.TagResponse, 0)

	req, err := http.NewRequest("DELETE", "/deleteDockerRepository?id=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteDockerRepository)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	dockerDAO.AssertNumberOfCalls(t, "DeleteDockerRepo", 1)

}
