package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockAudit "github.com/softplan/tenkai-api/pkg/audit/mocks"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
)

func doTest(t *testing.T, mode string) {

	appContext := AppContext{}

	mockEnvDao := mockEnvDaoWithLotOfThings(&appContext)
	mockConventionInterface(&appContext)

	mockVariableDAO := mockVariableDAOWithLotOfThings(&appContext)

	mockHelmSvc := mockHelmSvcWithLotOfThings(&appContext)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(getCharts())

	chartValue := `{"app":{"myvar":"myvalue"}}`
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(chartValue), nil)

	auditSvc := &mockAudit.AuditingInterface{}
	auditSvc.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.Repositories.VariableDAO = mockVariableDAO
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = auditSvc
	appContext.RabbitImpl = getMockRabbitMQ()

	user := mockUser()
	mockUserDAO := &mockRepo.UserDAOInterface{}
	mockUserDAO.On("FindByEmail", mock.Anything).Return(user, nil)
	appContext.Repositories.UserDAO = mockUserDAO

	mockDeploymentDAO := &mockRepo.DeploymentDAOInterface{}
	mockDeploymentDAO.On("CreateDeployment", mock.Anything).Return(1, nil)
	appContext.Repositories.DeploymentDAO = mockDeploymentDAO

	url := "/promote?mode=" + mode + "&srcEnvID=91&targetEnvID=92"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.promote)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")

}

func TestPromoteFull(t *testing.T) {
	doTest(t, "full")
}

func TestPromotePartial(t *testing.T) {
	doTest(t, "partial")
}

func TestPromote_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	url := "/promote?mode=full&srcEnvID=91&targetEnvID=92"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.promote)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be unauthorized.")
}

func doTestParamsError(t *testing.T, url string) {

	appContext := AppContext{}

	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.promote)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Mode missing")

}

func TestPromote_WithoutMode(t *testing.T) {
	doTestParamsError(t, "/promote?srcEnvID=91&targetEnvID=92")
}

func TestPromote_WithoutSrcEnvID(t *testing.T) {
	doTestParamsError(t, "/promote?mode=full&targetEnvID=92")
}

func TestPromote_WithoutTargetEnvID(t *testing.T) {
	doTestParamsError(t, "/promote?mode=full&srcEnvID=91")
}
