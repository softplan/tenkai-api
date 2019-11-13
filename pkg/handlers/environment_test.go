package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddEnvironments(t *testing.T) {

	var payload model.DataElement
	payload.Data.Group = "Test"
	payload.Data.Name = "Alfa"
	payload.Data.Namespace = "Beta"
	payload.Data.Gateway = "Tetra"
	payload.Data.CACertificate = "XPTOXPTOXPTO"
	payload.Data.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"

	payS, _ := json.Marshal(payload)

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockObject := &mocks.EnvironmentDAOInterface{}
	mockObject.On("CreateEnvironment", mock.Anything).Return(1, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockObject

	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/environments", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	roles := []string{"tenkai-admin"}
	principal := model.Principal{Name: "alfa", Email: "beta", Roles: roles}

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

func TestDeleteEnvironment(t *testing.T) {
	appContext := AppContext{}

	env := mockGetEnv()
	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("DeleteEnvironment", env).Return(nil)

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "DeleteEnvironment", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestDeleteEnvironment_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response is not Ok.")
}

func TestDeleteEnvironment_StringConvError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/environments/delete/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteEnvironment_GetByIDError(t *testing.T) {
	appContext := AppContext{}

	mockEnvDAO := mockGetByIDError(&appContext)

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteEnvironment_DeleteEnvironmentError(t *testing.T) {
	appContext := AppContext{}

	env := mockGetEnv()
	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("DeleteEnvironment", env).Return(errors.New("some error"))

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "DeleteEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditEnvironment(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("EditEnvironment", mock.Anything).Return(nil)

	var p model.DataElement
	env := mockGetEnv()
	env.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"
	p.Data = env

	req, err := http.NewRequest("POST", "/environments/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "EditEnvironment", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestEditEnvironment_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("POST", "/environments/edit", payload(mockGetEnv()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestEditEnvironment_UnmarshalPayload(t *testing.T) {
	appContext := AppContext{}
	req, err := http.NewRequest("POST", "/environments/edit", bytes.NewBuffer([]byte(`["invalid": 123]`)))
	assert.NoError(t, err)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditEnvironment_GetByIDError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByIDError(&appContext)

	var p model.DataElement
	env := mockGetEnv()
	env.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"
	p.Data = env

	req, err := http.NewRequest("POST", "/environments/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditEnvironment_EditEnvironmentError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("EditEnvironment", mock.Anything).Return(errors.New("some error"))

	var p model.DataElement
	env := mockGetEnv()
	env.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"
	p.Data = env

	req, err := http.NewRequest("POST", "/environments/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "EditEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDuplicateEnvironments(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("CreateEnvironment", mock.Anything).Return(1, nil)

	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, true, nil)

	appContext.Repositories.VariableDAO = mockVariableDAO

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	mockPrincipal(req, "tenkai-admin")

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "CreateEnvironment", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "CreateVariable", 2)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created.")
}

func TestDuplicateEnvironments_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestDuplicateEnvironments_StringConvError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/environments/duplicate/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDuplicateEnvironments_GetByIDError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByIDError(&appContext)

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDuplicateEnvironments_GetAllVarByEnvError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironmentError(&appContext)

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDuplicateEnvironments_CreateEnvError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)
	mockEnvDAO.On("CreateEnvironment", mock.Anything).Return(0, errors.New("some error"))

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "CreateEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDuplicateEnvironments_CreateVarError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)
	mockEnvDAO.On("CreateEnvironment", mock.Anything).Return(1, nil)
	mockVariableDAO.On("CreateVariable", mock.Anything).Return(nil, true, errors.New("some error"))

	req, err := http.NewRequest("GET", "/environments/duplicate/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-admin")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "CreateEnvironment", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "CreateVariable", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetEnvironments(t *testing.T) {
	appContext := AppContext{}
	mockEnvDAO := mockGetAllEnvironments(&appContext)

	req, err := http.NewRequest("GET", "/environments", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getEnvironments)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"Envs":[{"ID":999`)
	assert.Contains(t, response, `"group":"foo","name":"bar"`)
	assert.Contains(t, response, `"cluster_uri":"https://rancher-k8s-my-domain.com/k8s/clusters/c-kbfxr"`)
	assert.Contains(t, response, `"ca_certificate":"my-certificate"`)
	assert.Contains(t, response, `"token":"kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"`)
	assert.Contains(t, response, `"namespace":"dev","gateway":"my-gateway.istio-system.svc.cluster.local"`)
	assert.Contains(t, response, `"productVersion":""}]}`)
}

func TestGetEnvironments_AccessDenied(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/environments", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getEnvironments)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Response should be 405.")
}

func TestGetEnvironments_GetAllEnvError(t *testing.T) {
	appContext := AppContext{}
	mockEnvDAO := mockGetAllEnvironmentsError(&appContext)

	req, err := http.NewRequest("GET", "/environments", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getEnvironments)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestExport(t *testing.T) {
	appContext := AppContext{}
	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)

	req, err := http.NewRequest("GET", "/environments/export/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/export/{id}", appContext.export).Methods("GET")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `global username=user`)
	assert.Contains(t, response, `bar password=password`)
}

func TestExport_GetAllVarByEnvError(t *testing.T) {
	appContext := AppContext{}
	mockVariableDAO := mockGetAllVariablesByEnvironmentError(&appContext)

	req, err := http.NewRequest("GET", "/environments/export/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/export/{id}", appContext.export).Methods("GET")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestExport_StringConvError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/environments/export/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/export/{id}", appContext.export).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetAllEnvironments(t *testing.T) {
	appContext := AppContext{}
	mockEnvDAO := mockGetAllEnvironmentsPrincipal(&appContext, "")

	req, err := http.NewRequest("GET", "/environments/all", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getAllEnvironments)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"Envs":[{"ID":999`)
	assert.Contains(t, response, `"group":"foo","name":"bar"`)
	assert.Contains(t, response, `"cluster_uri":"https://rancher-k8s-my-domain.com/k8s/clusters/c-kbfxr"`)
	assert.Contains(t, response, `"ca_certificate":"my-certificate"`)
	assert.Contains(t, response, `"token":"kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"`)
	assert.Contains(t, response, `"namespace":"dev","gateway":"my-gateway.istio-system.svc.cluster.local"`)
	assert.Contains(t, response, `"productVersion":""}]}`)
}

func TestGetAllEnvironments_GetAllEnvError(t *testing.T) {
	appContext := AppContext{}
	mockEnvDAO := mockGetAllEnvironmentsError(&appContext)

	req, err := http.NewRequest("GET", "/environments/all", bytes.NewBuffer(nil))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getAllEnvironments)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}
