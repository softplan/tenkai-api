package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"encoding/json"

	mockAud "github.com/softplan/tenkai-api/pkg/audit/mocks"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCharts(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockObject := &mockSvc.HelmServiceInterface{}

	data := make([]model.SearchResult, 1)
	data[0].Name = "test-chart"
	data[0].ChartVersion = "1.0"
	data[0].Description = "Test only"
	data[0].AppVersion = "1.0"

	mockObject.On("SearchCharts", mock.Anything, mock.Anything).Return(&data)
	appContext.HelmServiceAPI = mockObject

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

	assert.Equal(t, getExpect(data), string(rr.Body.Bytes()), "Response is not correct.")
}

func TestDeleteHelmRelease(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("DELETE", "/deleteHelmRelease?environmentID=999&releaseName=foo&purge=false", nil)
	assert.Nil(t, err, "Request err should be nil.")

	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := model.Environment{Group: "foo", Name: "bar"}
	env.ID = 999
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)
	envs := []model.Environment{env}
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(envs, nil)

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("DeleteHelmRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockAudit := mockAud.AuditingInterface{}
	mockAudit.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = &mockAudit

	roles := []string{"tenkai-user"}
	principal := model.Principal{Name: "alfa", Email: "beta", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.deleteHelmRelease)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func getExpect(sr []model.SearchResult) string {
	j, _ := json.Marshal(sr)
	return "{\"chart\":" + string(j) + "}"
}
