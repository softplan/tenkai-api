package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"
	"github.com/softplan/tenkai-api/pkg/audit"
	"github.com/softplan/tenkai-api/pkg/configs"
	"github.com/softplan/tenkai-api/pkg/dbms"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/service/core"
	dockerapi "github.com/softplan/tenkai-api/pkg/service/docker"
	"log"
	"net/http"
	"strings"
	"sync"
)

//Repositories  Repositories
type Repositories struct {
	ConfigDAO              repository.ConfigDAOInterface
	DockerDAO              repository.DockerDAOInterface
	EnvironmentDAO         repository.EnvironmentDAOInterface
	ProductDAO             repository.ProductDAOInterface
	SolutionDAO            repository.SolutionDAOInterface
	SolutionChartDAO       repository.SolutionChartDAOInterface
	UserDAO                repository.UserDAOInterface
	VariableDAO            repository.VariableDAOInterface
	ValueRuleDAO           repository.ValueRuleDAOInterface
	VariableRuleDAO        repository.VariableRuleDAOInterface
	CompareEnvsQueryDAO    repository.CompareEnvsQueryDAOInterface
	SecurityOperationDAO   repository.SecurityOperationDAOInterface
	UserEnvironmentRoleDAO repository.UserEnvironmentRoleDAOInterface
}

//AppContext AppContext
type AppContext struct {
	ConventionInterface core.ConventionInterface
	DockerServiceAPI    dockerapi.DockerServiceInterface
	HelmServiceAPI      helmapi.HelmServiceInterface
	Auditing            audit.AuditingInterface
	K8sConfigPath       string
	Configuration       *configs.Configuration
	Repositories        Repositories
	Database            dbms.Database
	Elk                 *elastic.Client
	Mutex               sync.Mutex
	ChartImageCache     sync.Map
	DockerTagsCache     sync.Map
	ConfigMapCache      sync.Map
}

func defineRotes(r *mux.Router, appContext *AppContext) {

	r.HandleFunc("/getVirtualServices", appContext.getVirtualServices).Methods("GET")
	r.HandleFunc("/install", appContext.install).Methods("POST")
	r.HandleFunc("/multipleInstall", appContext.multipleInstall).Methods("POST")
	r.HandleFunc("/getHelmCommand", appContext.getHelmCommand).Methods("POST")

	r.HandleFunc("/getVariablesNotUsed/{id}", appContext.getVariablesNotUsed).Methods("GET")

	r.HandleFunc("/listVariables", appContext.getVariablesByEnvironmentAndScope).Methods("POST")
	r.HandleFunc("/saveVariableValues", appContext.saveVariableValues).Methods("POST")
	r.HandleFunc("/getChartVariables", appContext.getChartVariables).Methods("POST")
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.HandleFunc("/listReleaseHistory", appContext.listReleaseHistory).Methods("POST")
	r.HandleFunc("/rollback", appContext.rollback).Methods("POST")

	r.HandleFunc("/charts/{repo}", appContext.listCharts).Methods("GET")
	r.HandleFunc("/listPods/{id}", appContext.pods).Methods("GET")
	r.HandleFunc("/listServices/{id}", appContext.services).Methods("GET")

	r.HandleFunc("/variables", appContext.editVariable).Methods("POST")
	r.HandleFunc("/variables/copy-value", appContext.copyVariableValue).Methods("POST")
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.HandleFunc("/deletePod", appContext.deletePod).Methods("DELETE")

	r.HandleFunc("/variables/edit", appContext.editVariable).Methods("POST")

	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.HandleFunc("/environments/edit", appContext.editEnvironment).Methods("POST")
	r.HandleFunc("/environments", appContext.addEnvironments).Methods("POST")
	r.HandleFunc("/environments", appContext.getEnvironments).Methods("GET")
	r.HandleFunc("/environments/all", appContext.getAllEnvironments).Methods("GET")
	r.HandleFunc("/environments/export/{id}", appContext.export).Methods("GET")
	r.HandleFunc("/hasConfigMap", appContext.hasConfigMap).Methods("POST")

	r.HandleFunc("/revision", appContext.revision).Methods("POST")

	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")

	r.HandleFunc("/repositories", appContext.listRepositories).Methods("GET")
	r.HandleFunc("/repositories", appContext.newRepository).Methods("POST")
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")

	r.HandleFunc("/deleteHelmRelease", appContext.deleteHelmRelease).Methods("DELETE")
	r.HandleFunc("/helmDryRun", appContext.helmDryRun).Methods("POST")

	r.HandleFunc("/solutions", appContext.listSolution).Methods("GET")
	r.HandleFunc("/solutions", appContext.newSolution).Methods("POST")
	r.HandleFunc("/solutions/edit", appContext.editSolution).Methods("POST")
	r.HandleFunc("/solutions/{id}", appContext.deleteSolution).Methods("DELETE")

	r.HandleFunc("/products", appContext.listProducts).Methods("GET")
	r.HandleFunc("/products", appContext.newProduct).Methods("POST")
	r.HandleFunc("/products/edit", appContext.editProduct).Methods("POST")
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")

	r.HandleFunc("/productVersions", appContext.listProductVersions).Methods("GET")
	r.HandleFunc("/productVersions", appContext.newProductVersion).Methods("POST")
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.HandleFunc("/productVersions/lock/{id}", appContext.lockProductVersion).Methods("GET")
	r.HandleFunc("/productVersions/unlock/{id}", appContext.unlockProductVersion).Methods("GET")

	r.HandleFunc("/productVersionServices", appContext.listProductVersionServices).Methods("GET")
	r.HandleFunc("/productVersionServices", appContext.newProductVersionService).Methods("POST")
	r.HandleFunc("/productVersionServices/edit", appContext.editProductVersionService).Methods("POST")
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")

	r.HandleFunc("/dockerRepo", appContext.listDockerRepositories).Methods("GET")
	r.HandleFunc("/dockerRepo", appContext.newDockerRepository).Methods("POST")
	r.HandleFunc("/dockerRepo/{id}", appContext.deleteDockerRepository).Methods("DELETE")

	r.HandleFunc("/solutionCharts", appContext.listSolutionCharts).Methods("GET")
	r.HandleFunc("/solutionCharts", appContext.newSolutionChart).Methods("POST")
	r.HandleFunc("/solutionCharts/{id}", appContext.deleteSolutionChart).Methods("DELETE")

	r.HandleFunc("/deployTrafficRule", appContext.deployTrafficRule).Methods("POST")

	r.HandleFunc("/repoUpdate", appContext.repoUpdate).Methods("GET")

	r.HandleFunc("/repo/default", appContext.setDefaultRepo).Methods("POST")
	r.HandleFunc("/repo/default", appContext.getDefaultRepo).Methods("GET")

	r.HandleFunc("/users/createOrUpdate", appContext.createOrUpdateUser).Methods("POST")

	r.HandleFunc("/users", appContext.newUser).Methods("POST")
	r.HandleFunc("/users", appContext.listUsers).Methods("GET")
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")

	r.HandleFunc("/promote", appContext.promote).Methods("GET")

	r.HandleFunc("/listDockerTags", appContext.listDockerTags).Methods("POST")

	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}",
		appContext.newEnvironmentPermission).Methods("GET")

	r.HandleFunc("/settings", appContext.addSettings).Methods("POST")
	r.HandleFunc("/getSettingList", appContext.getSettingList).Methods("POST")

	r.HandleFunc("/valuerules", appContext.listValueRules).Methods("GET")
	r.HandleFunc("/valuerules", appContext.newValueRule).Methods("POST")
	r.HandleFunc("/valuerules/edit", appContext.editValueRule).Methods("POST")
	r.HandleFunc("/valuerules/{id}", appContext.deleteValueRule).Methods("DELETE")

	r.HandleFunc("/variablerules", appContext.listVariableRules).Methods("GET")
	r.HandleFunc("/variablerules", appContext.newVariableRule).Methods("POST")
	r.HandleFunc("/variablerules/edit", appContext.editVariableRule).Methods("POST")
	r.HandleFunc("/variablerules/{id}", appContext.deleteVariableRule).Methods("DELETE")

	r.HandleFunc("/validateVariables", appContext.validateVariables).Methods("POST")
	r.HandleFunc("/validateEnvVars/{envId}", appContext.validateEnvironmentVariables).Methods("POST")

	r.HandleFunc("/compare-environments", appContext.compareEnvironments).Methods("POST")
	r.HandleFunc("/compare-environments/save-query", appContext.saveCompareEnvQuery).Methods("POST")
	r.HandleFunc("/compare-environments/load-queries", appContext.loadCompareEnvQueries).Methods("GET")
	r.HandleFunc("/compare-environments/delete-query/{id}", appContext.deleteCompareEnvQuery).Methods("DELETE")

	r.HandleFunc("/security-operations", appContext.listSecurityOperation).Methods("GET")
	r.HandleFunc("/security-operations", appContext.createOrUpdateSecurityOperation).Methods("POST")
	r.HandleFunc("/security-operations/{id}", appContext.deleteSecurityOperation).Methods("DELETE")

	r.HandleFunc("/getUserPolicyByEnvironment", appContext.getUserPolicyByEnvironment).Methods("POST")
	r.HandleFunc("/createOrUpdateUserEnvironmentRole", appContext.createOrUpdateUserEnvironmentRole).Methods("POST")

	r.HandleFunc("/", appContext.rootHandler)

}

//StartHTTPServer StartHTTPServer
func StartHTTPServer(appContext *AppContext) {

	port := appContext.Configuration.Server.Port
	global.Logger.Info(global.AppFields{global.Function: "startHTTPServer", "port": port}, "online - listen and server")

	r := mux.NewRouter()

	defineRotes(r, appContext)

	log.Fatal(http.ListenAndServe(":"+port, commonHandler(r)))

}

func extractToken(reqToken string) *model.Principal {
	var principal model.Principal

	splitToken := strings.Split(reqToken, "Bearer ")
	reqToken = splitToken[1]
	token, _, earl := new(jwt.Parser).ParseUnverified(reqToken, jwt.MapClaims{})
	if earl == nil {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			in := claims["realm_access"]
			if in != nil {
				realmAccessMap := in.(map[string]interface{})
				roles := realmAccessMap["roles"]
				if roles != nil {
					elements := roles.([]interface{})
					for _, element := range elements {
						principal.Roles = append(principal.Roles, element.(string))
					}
				}
			}
			principal.Name = fmt.Sprintf("%v", claims["name"])
			principal.Email = fmt.Sprintf("%v", claims["email"])
			return &principal
		}
	}
	return nil
}

func commonHandler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		reqToken := r.Header.Get("Authorization")
		if len(reqToken) > 0 {
			principal := extractToken(reqToken)
			if principal != nil {
				data, _ := json.Marshal(*principal)
				r.Header.Set("principal", string(data))
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (appContext *AppContext) hasAccess(email string, envID int) (bool, error) {
	result := false
	environments, err := appContext.Repositories.EnvironmentDAO.GetAllEnvironments(email)
	if err != nil {
		return false, err
	}
	for _, e := range environments {
		if e.ID == uint(envID) {
			result = true
			break
		}
	}
	return result, nil
}

func (appContext *AppContext) hasEnvironmentRole(principal model.Principal, envID uint, role string) (bool, error) {
	var user model.User
	var err error
	if user, err = appContext.Repositories.UserDAO.FindByEmail(principal.Email); err != nil {
		return false, err
	}
	result := &model.SecurityOperation{}
	if result, err = appContext.Repositories.UserEnvironmentRoleDAO.
		GetRoleByUserAndEnvironment(user, envID); err != nil {
		return false, err
	}
	authorized := false
	if result != nil {

		for _, e := range result.Policies {
			if e == role {
				authorized = true
				break
			}
		}
	}
	return authorized, nil
}

func (appContext *AppContext) rootHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"service": "TENKAI",
		"status":  "ready",
	}

	json, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set(global.ContentType, "application/json")
	w.Write(json)
}
