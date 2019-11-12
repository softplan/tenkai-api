package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetVirtualServices(t *testing.T) {

	appContext := AppContext{}

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := mockGetByID(&appContext)
	mockEnvDao.On("GetAllEnvironments", "beta@alfa.com").Return(envs, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao

	mockHelmSvc := &mocks.HelmServiceInterface{}

	result := make([]string, 1)
	result = append(result, "test.com.br")
	mockHelmSvc.On("GetVirtualServices", mock.Anything, mock.Anything).Return(result, nil)

	mockConvention := mockConventionInterface(&appContext)

	appContext.HelmServiceAPI = mockHelmSvc

	req, err := http.NewRequest("GET", "/getVirtualServices?environmentID=999", bytes.NewBuffer(nil))
	assert.NoError(t, err)

	mockPrincipal(req, "tenkai-helm-upgrade")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getVirtualServices)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Created.")

	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

}
