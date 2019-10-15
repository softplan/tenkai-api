package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/stretchr/testify/mock"
)

type EnvironmentDAOMock struct {
	mock.Mock
}

func (EnvironmentDAOMock) CreateEnvironment(env model.Environment) (int, error) {
	return 1, nil
}

func (EnvironmentDAOMock) EditEnvironment(env model.Environment) error {
	return nil
}

func (EnvironmentDAOMock) DeleteEnvironment(env model.Environment) error {
	return nil
}

func (EnvironmentDAOMock) GetAllEnvironments(principal string) ([]model.Environment, error) {
	return make([]model.Environment, 0), nil
}

func (EnvironmentDAOMock) GetByID(envID int) (*model.Environment, error) {
	return &model.Environment{}, nil
}

func TestAddEnvironments(t *testing.T) {

	var payload model.DataElement
	payload.Data.Group = "Test"
	payload.Data.Name = "Alfa"
	payload.Data.Namespace = "Beta"
	payload.Data.Gateway = "Tetra"
	payload.Data.CACertificate = "XPTOXPTOXPTO"
	payload.Data.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"

	payS, _ := json.Marshal(payload)

	appContext := appContext{}
	appContext.k8sConfigPath = "/tmp/"
	appContext.environmentDAO = &EnvironmentDAOMock{}

	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/environments", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	roles := []string{"tenkai-admin"}
	principal := model.Principal{Name:"alfa", Email:"beta", Roles:roles}

	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.addEnvironments)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}


}