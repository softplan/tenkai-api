package handlers

import (
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	mockAudit "github.com/softplan/tenkai-api/pkg/audit/mocks"
)

func TestPromoteFull(t *testing.T) {

	appContext := AppContext{}

	var envs []model.Environment
	envs = append(envs, mockGetEnv())
	mockEnvDao := &mocks.EnvironmentDAOInterface{}

	env := mockGetEnv()
	mockEnvDao.On("GetByID", mock.Anything).Return(&env, nil)
	mockEnvDao.On("GetAllEnvironments", mock.Anything).Return(envs, nil)
	mockConvention := mockConventionInterface(&appContext)

	mockVariableDAO := mockGetAllVariablesByEnvironmentAndScope(&appContext)
	var variables []model.Variable
	variable := mockGlobalVariable()
	variables = append(variables, variable)
	mockVariableDAO.On("DeleteVariableByEnvironmentID", mock.Anything).Return(nil)
	mockVariableDAO.On("GetAllVariablesByEnvironment", mock.Anything).Return(variables, nil)
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, true, nil)

	hlr := helmapi.HelmListResult{}
	hlr.Releases = make([]helmapi.ListRelease, 0)

	lr := helmapi.ListRelease{}
	lr.Name = "tjusuarios-master"
	lr.Chart = "tjusuarios-master"
	lr.Namespace = "master"
	lr.Status = "Running"
	lr.AppVersion = "1.0"
	lr.Revision = 1

	hlr.Releases = append(hlr.Releases, lr)

	mockHelmSvc := mockUpgrade(&appContext)
	mockHelmSvc.On("ListHelmDeployments", mock.Anything, mock.Anything).Return(&hlr, nil)
	mockHelmSvc.On("DeleteHelmRelease", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	auditSvc := &mockAudit.AuditingInterface{}
	auditSvc.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.Repositories.VariableDAO = mockVariableDAO
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = auditSvc

	req, err := http.NewRequest("GET", "/promote?mode=full&srcEnvID=91&targetEnvID=92", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{constraints.TenkaiPromote})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.promote)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "DeleteVariableByEnvironmentID", 1)

}
