package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddSettings(t *testing.T) {

	var payload model.SettingsList
	payload.List = make([]model.Settings, 0)

	element := model.Settings{}
	element.Name = "repo1"
	element.Value = "repovalue"

	payload.List = append(payload.List, element)

	payS, _ := json.Marshal(payload)

	appContext := AppContext{}

	configDAO := mocks.ConfigDAOInterface{}
	configDAO.On("CreateOrUpdateConfig", mock.Anything).Return(1, nil)
	appContext.Repositories.ConfigDAO = &configDAO

	req, err := http.NewRequest("POST", "/settings", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.addSettings)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	configDAO.AssertNumberOfCalls(t, "CreateOrUpdateConfig", 1)

}

func TestGetSettingList(t *testing.T) {

	var payload model.GetSettingsListRequest
	payload.List = make([]string, 0)
	payload.List = append(payload.List, "element-1")

	payS, _ := json.Marshal(payload)

	result := model.ConfigMap{}
	result.Value = "abc"
	result.Name = "xpto"

	appContext := AppContext{}

	configDAO := mocks.ConfigDAOInterface{}
	configDAO.On("GetConfigByName", mock.Anything).Return(result, nil)
	appContext.Repositories.ConfigDAO = &configDAO

	req, err := http.NewRequest("POST", "/getSettingList", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getSettingList)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	configDAO.AssertNumberOfCalls(t, "GetConfigByName", 1)

}
