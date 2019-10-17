package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
)

func (appContext *AppContext) listCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	vars := mux.Vars(r)
	repo := vars["repo"]

	all, ok := r.URL.Query()["all"]
	allVersions := true
	if ok && len(all[0]) > 0 {
		allVersions = all[0] == "true"
	}

	searchTerms := []string{repo}
	searchResult := appContext.HelmServiceAPI.SearchCharts(searchTerms, allVersions)
	result := &model.ChartsResult{Charts: *searchResult}

	data, _ := json.Marshal(result)

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deleteHelmRelease(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		http.Error(w, errors.New("param environmentID is required").Error(), 501)
		return
	}

	releasesName, ok := r.URL.Query()["releaseName"]
	if !ok || len(releasesName[0]) < 1 {
		http.Error(w, errors.New("param releasesName is required").Error(), 501)
		return
	}

	purges, ok := r.URL.Query()["purge"]
	if !ok || len(purges[0]) < 1 {
		http.Error(w, errors.New("param purges, is required").Error(), 501)
		return
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(envID))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	purge, _ := strconv.ParseBool(purges[0])
	err = appContext.HelmServiceAPI.DeleteHelmRelease(kubeConfig, releasesName[0], purge)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	auditValues := make(map[string]string)
	auditValues["environment"] = environment.Name
	auditValues["purge"] = strconv.FormatBool(purge)
	auditValues["name"] = releasesName[0]

	appContext.Auditory.DoAudit(r.Context(), appContext.Elk, principal.Email, "deleteHelmRelease", auditValues)

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) rollback(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	err = appContext.HelmServiceAPI.RollbackRelease(kubeConfig, payload.ReleaseName, payload.Revision)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) revision(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	yaml, err := appContext.HelmServiceAPI.Get(kubeConfig, payload.ReleaseName, payload.Revision)

	data, _ := json.Marshal(yaml)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listReleaseHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.HistoryRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	history, err := appContext.HelmServiceAPI.GetHelmReleaseHistory(kubeConfig, payload.ReleaseName)

	data, _ := json.Marshal(history)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listHelmDeploymentsByEnvironment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

	result, err := appContext.HelmServiceAPI.ListHelmDeployments(kubeConfig, environment.Namespace)

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) hasConfigMap(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, payload.ChartName, payload.ChartVersion, "deployment")

	w.WriteHeader(http.StatusOK)
	if err != nil {
		w.Write([]byte("{\"result\":\"false\"}"))
	} else {
		deployment := string(result)
		if strings.Index(deployment, "gcm") > 0 {
			w.Write([]byte("{\"result\":\"true\"}"))
		} else {
			w.Write([]byte("{\"result\":\"false\"}"))
		}
	}

}

func (appContext *AppContext) getChartVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, payload.ChartName, payload.ChartVersion, "values")

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (appContext *AppContext) getHelmCommand(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiHelmUpgrade) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	var fullCommand string
	for _, element := range payload.Deployables {

		//Locate Environment
		environment, err := appContext.Repositories.EnvironmentDAO.GetByID(element.EnvironmentID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		command, errX := appContext.simpleInstall(environment, element.Chart, element.ChartVersion, element.Name, out, false, true)
		if errX != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		fullCommand = fullCommand + "\n\n" + command

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fullCommand))

}

func (appContext *AppContext) multipleInstall(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiHelmUpgrade) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	for _, element := range payload.Deployables {

		//Locate Environment
		environment, err := appContext.Repositories.EnvironmentDAO.GetByID(element.EnvironmentID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		_, err = appContext.simpleInstall(environment, element.Chart, element.ChartVersion, element.Name, out, false, false)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		auditValues := make(map[string]string)
		auditValues["environment"] = environment.Name
		auditValues["chartName"] = element.Chart
		auditValues["name"] = element.Name

		appContext.Auditory.DoAudit(r.Context(), appContext.Elk, principal.Email, "deploy", auditValues)

	}

	fmt.Println(out.String())

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) install(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiHelmUpgrade) {
		http.Error(w, errors.New("Access Denied").Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//TODO Verify if chart exists
	_, err = appContext.simpleInstall(environment, payload.Chart, payload.ChartVersion, payload.Name, out, false, false)
	if err != nil {
		fmt.Println(out.String())
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) helmDryRun(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//TODO Verify if chart exists
	_, err = appContext.simpleInstall(environment, payload.Chart, payload.ChartVersion, payload.Name, out, true, false)

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())

}

func (appContext *AppContext) simpleInstall(environment *model.Environment, chart string, chartVersion string,
	name string, out *bytes.Buffer, dryRun bool, helmCommandOnly bool) (string, error) {

	//TODO - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.

	searchTerm := chart
	if strings.Index(name, "gcm") > -1 {
		searchTerm = name
	}
	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(int(environment.ID), searchTerm)
	globalVariables := appContext.getGlobalVariables(int(environment.ID))

	var args []string
	for i, item := range variables {

		if item.Secret {
			byteValues, _ := hex.DecodeString(item.Value)
			value, err := util.Decrypt(byteValues, appContext.Configuration.App.Passkey)
			if err == nil {
				variables[i].Value = string(value)
			}
		}

		if len(item.Name) > 0 && len(item.Value) > 0 {
			args = append(args, normalizeVariableName(item.Name)+"="+replace(item.Value, *environment, globalVariables))
		}

	}

	//Add Default Gateway
	if len(environment.Gateway) > 0 {
		args = append(args, "istio.virtualservices.gateways[0]="+environment.Gateway)
	}

	dt := time.Now()
	args = append(args, "app.dateHour="+dt.String())

	if err == nil {
		name := name + "-" + environment.Namespace
		kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name

		if !helmCommandOnly {
			err := appContext.HelmServiceAPI.Upgrade(kubeConfig, name, chart, chartVersion, environment.Namespace, args, out, dryRun)
			if err != nil {
				return "", err
			}
		} else {
			var message string

			message = "helm upgrade --install " + name + " \\\n"

			for _, e := range args {
				message = message + " --set \"" + e + "\" " + " \\\n"
			}

			message = message + " " + chart + " --namespace=" + environment.Namespace

			return message, nil
		}
	}

	return "", nil
}

func replace(value string, environment model.Environment, variables []model.Variable) string {
	newValue := strings.Replace(value, "${NAMESPACE}", environment.Namespace, -1)
	keywords := util.GetReplacebleKeyName(newValue)
	for _, keyword := range keywords {
		for _, element := range variables {
			if element.Name == keyword {
				newValue = strings.Replace(newValue, "${"+element.Name+"}", element.Value, -1)
				break
			}
		}
	}
	return newValue
}

func normalizeVariableName(value string) string {
	if strings.Index(value, "istio.") > -1 || (strings.Index(value, "image.")) > -1 || (strings.Index(value, "service.")) > -1 {
		return value
	}
	return "app." + value
}

func (appContext *AppContext) getGlobalVariables(id int) []model.Variable {
	variables, _ := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(id, "global")

	for i, e := range variables {
		if e.Secret {
			byteValues, _ := hex.DecodeString(e.Value)
			value, err := util.Decrypt(byteValues, appContext.Configuration.App.Passkey)
			if err == nil {
				variables[i].Value = string(value)
			}
		}
	}

	return variables
}
