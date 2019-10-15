package main

import (
	"bytes"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/service/helm/mocks"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCharts(t *testing.T) {

	appContext := appContext{}
	appContext.k8sConfigPath = "/tmp/"

	mockObject := &mocks.HelmServiceInterface{}

	data := make([]model.SearchResult, 1)
	data[0].Name = "test-chart"
	data[0].ChartVersion = "1.0"
	data[0].Description = "Test only"
	data[0].AppVersion = "1.0"

	mockObject.On("SearchCharts", mock.Anything, mock.Anything).Return(&data)
	appContext.helmServiceAPI = mockObject

	req, err := http.NewRequest("GET", "/listCharts", bytes.NewBuffer(nil))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listCharts)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}
