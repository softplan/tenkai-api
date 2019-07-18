package main

import (
	"github.com/softplan/tenkai-api/dbms"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"log"
	"net/http"
	"os"

	"github.com/softplan/tenkai-api/configs"
	"github.com/softplan/tenkai-api/global"

	"github.com/gorilla/mux"
)

const (
	configFileName = "app"
)

type appContext struct {
	configuration *configs.Configuration
	database      dbms.Database
}

func main() {
	logFields := global.AppFields{global.FUNCTION: "main"}

	_ = os.Mkdir(global.KUBECONFIG_BASE_PATH, 0777)

	if _, err := os.Stat(global.HELM_DIR + "/repository/repositories.yaml"); os.IsNotExist(err) {
		helmapi.InitializeHelm()
	}

	global.Logger.Info(logFields, "carregando configurações")

	config, error := configs.ReadConfig(configFileName)
	checkFatalError(error)

	appContext := &appContext{configuration: config}

	dbmsUri := config.App.Dbms.Uri

	//Conecta no postgres
	appContext.database.Connect(dbmsUri)
	defer appContext.database.Db.Close()

	global.Logger.Info(logFields, "iniciando o servidor http")
	startHTTPServer(appContext)
}

func startHTTPServer(appContext *appContext) {

	port := appContext.configuration.Server.Port
	global.Logger.Info(global.AppFields{global.FUNCTION: "startHTTPServer", "port": port}, "online - listen and server")

	r := mux.NewRouter()

	r.HandleFunc("/install", appContext.install).Methods("POST")
	r.HandleFunc("/multipleInstall", appContext.multipleInstall).Methods("POST")

	r.HandleFunc("/listVariables", appContext.getVariablesByEnvironmentAndScope).Methods("POST")
	r.HandleFunc("/saveVariableValues", appContext.saveVariableValues).Methods("POST")
	r.HandleFunc("/getChartVariables", appContext.getChartVariables).Methods("POST")
	r.HandleFunc("/listHelmDeployments", appContext.listHelmDeployments).Methods("GET")
	r.HandleFunc("/charts/{repo}", appContext.listCharts).Methods("GET")

	r.HandleFunc("/variables", appContext.addVariables).Methods("POST")
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.HandleFunc("/variables/edit", appContext.editVariable).Methods("POST")

	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.HandleFunc("/environments/edit", appContext.editEnvironment).Methods("POST")
	r.HandleFunc("/environments", appContext.addEnvironments).Methods("POST")
	r.HandleFunc("/environments", appContext.getEnvironments).Methods("GET")

	r.HandleFunc("/repositories", appContext.listRepositories).Methods("GET")
	r.HandleFunc("/repositories", appContext.newRepository).Methods("POST")
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")

	r.HandleFunc("/releases", appContext.listReleases).Methods("GET")
	r.HandleFunc("/releases", appContext.newRelease).Methods("POST")
	r.HandleFunc("/releases/{id}", appContext.deleteRelease).Methods("DELETE")

	r.HandleFunc("/dependencies", appContext.listDependencies).Methods("GET")
	r.HandleFunc("/dependencies", appContext.newDependency).Methods("POST")
	r.HandleFunc("/dependencies/{id}", appContext.deleteDependency).Methods("DELETE")

	r.HandleFunc("/analyse", appContext.analyse).Methods("POST")

	r.HandleFunc("/", appContext.rootHandler)

	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(r)))

}
