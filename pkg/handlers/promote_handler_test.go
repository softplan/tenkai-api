package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mockAudit "github.com/softplan/tenkai-api/pkg/audit/mocks"
)

func doTest(t *testing.T, mode string) {

	appContext := AppContext{}

	mockEnvDao := mockEnvDaoWithLotOfThings(&appContext)
	mockConventionInterface(&appContext)

	mockVariableDAO := mockVariableDAOWithLotOfThings(&appContext)

	mockHelmSvc := mockHelmSvcWithLotOfThings(&appContext)
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(getCharts())

	auditSvc := &mockAudit.AuditingInterface{}
	auditSvc.On("DoAudit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.Repositories.VariableDAO = mockVariableDAO
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Auditing = auditSvc

	url := "/promote?mode=" + mode + "&srcEnvID=91&targetEnvID=92"
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, constraints.TenkaiAdmin)

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

	mockPrincipal(req, "role-unauthorized")

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

	mockPrincipal(req, constraints.TenkaiAdmin)

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
