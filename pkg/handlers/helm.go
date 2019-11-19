package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/util"

	"strings"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

func (appContext *AppContext) listCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

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
		http.Error(w, errors.New("param environmentID is required").Error(), http.StatusInternalServerError)
		return
	}

	releasesName, ok := r.URL.Query()["releaseName"]
	if !ok || len(releasesName[0]) < 1 {
		http.Error(w, errors.New("param releasesName is required").Error(), http.StatusInternalServerError)
		return
	}

	purges, ok := r.URL.Query()["purge"]
	if !ok || len(purges[0]) < 1 {
		http.Error(w, errors.New("param purges, is required").Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	purge, _ := strconv.ParseBool(purges[0])
	err = appContext.HelmServiceAPI.DeleteHelmRelease(kubeConfig, releasesName[0], purge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	auditValues := make(map[string]string)
	auditValues["environment"] = environment.Name
	auditValues["purge"] = strconv.FormatBool(purge)
	auditValues["name"] = releasesName[0]

	appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "deleteHelmRelease", auditValues)

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) rollback(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	err = appContext.HelmServiceAPI.RollbackRelease(kubeConfig, payload.ReleaseName, payload.Revision)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) revision(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	yaml, err := appContext.HelmServiceAPI.Get(kubeConfig, payload.ReleaseName, payload.Revision)

	data, _ := json.Marshal(yaml)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listReleaseHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.HistoryRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	history, err := appContext.HelmServiceAPI.GetHelmReleaseHistory(kubeConfig, payload.ReleaseName)

	data, _ := json.Marshal(history)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listHelmDeploymentsByEnvironment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(global.ContentType, global.JSONContentType)

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	result, err := appContext.HelmServiceAPI.ListHelmDeployments(kubeConfig, environment.Namespace)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) hasConfigMap(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

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

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chartName, err := appContext.getChartName(payload.ChartName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, chartName, payload.ChartVersion, "values")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (appContext *AppContext) getHelmCommand(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiHelmUpgrade) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := &bytes.Buffer{}

	var fullCommand string
	for _, element := range payload.Deployables {

		//Locate Environment
		environment, err := appContext.Repositories.EnvironmentDAO.GetByID(element.EnvironmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		command, errX := appContext.simpleInstall(environment, element, out, false, true)
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
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		msg := err.Error()
		if err.Error() == "record not found" {
			msg = "Environment " + strconv.Itoa(payload.EnvironmentID) + " not found"
		}
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	for _, element := range payload.Deployables {

		_, err = appContext.simpleInstall(environment, element, out, false, false)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		auditValues := make(map[string]string)
		auditValues["environment"] = environment.Name
		auditValues["chartName"] = element.Chart
		auditValues["name"] = element.Name

		appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "deploy", auditValues)

	}

	if payload.ProductVersionID > 0 {
		pv, err := appContext.Repositories.ProductDAO.ListProductVersionsByID(payload.ProductVersionID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
		environment.ProductVersion = pv.Version
		if err := appContext.Repositories.EnvironmentDAO.EditEnvironment(*environment); err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) install(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiHelmUpgrade) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = appContext.simpleInstall(environment, payload, out, false, false)
	if err != nil {
		fmt.Println(out.String())
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) helmDryRun(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = appContext.simpleInstall(environment, payload, out, true, false)

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())

}

func (appContext *AppContext) getArgs(variables []model.Variable, globalVariables []model.Variable, environment *model.Environment) []string {

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
	return args
}

func (appContext *AppContext) simpleInstall(environment *model.Environment, installPayload model.InstallPayload, out *bytes.Buffer, dryRun bool, helmCommandOnly bool) (string, error) {

	//WARNING - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.

	searchTerm := installPayload.Chart
	if strings.Index(installPayload.Name, "gcm") > -1 {
		searchTerm = installPayload.Name
	}
	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(int(environment.ID), searchTerm)
	globalVariables := appContext.getGlobalVariables(int(environment.ID))

	args := appContext.getArgs(variables, globalVariables, environment)

	//Add Default Gateway
	if len(environment.Gateway) > 0 {
		args = append(args, "istio.virtualservices.gateways[0]="+environment.Gateway)
	}

	dt := time.Now()
	args = append(args, "app.dateHour="+dt.String())

	if err == nil {
		name := installPayload.Name + "-" + environment.Namespace
		kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

		if !helmCommandOnly {

			upgradeRequest := helmapi.UpgradeRequest{}
			upgradeRequest.Kubeconfig = kubeConfig
			upgradeRequest.Namespace = environment.Namespace
			upgradeRequest.ChartVersion = installPayload.ChartVersion
			upgradeRequest.Chart = installPayload.Chart
			upgradeRequest.Variables = args
			upgradeRequest.Dryrun = dryRun
			upgradeRequest.Release = name

			return appContext.doUpgrade(upgradeRequest, out)

		}

		return getHelmMessage(name, args, environment, installPayload.Chart), nil

	}

	return "", nil
}

func (appContext *AppContext) getChartName(name string) (string, error) {

	searchTerms := []string{name}
	searchResult := appContext.HelmServiceAPI.SearchCharts(searchTerms, false)

	if len(*searchResult) > 0 {
		r := *searchResult
		return r[0].Name, nil
	}
	return "", errors.New("Chart does not exists")
}

func (appContext *AppContext) doUpgrade(upgradeRequest helmapi.UpgradeRequest, out *bytes.Buffer) (string, error) {
	var err error
	upgradeRequest.Chart, err = appContext.getChartName(upgradeRequest.Chart)
	if err != nil {
		return "", err
	}
	err = appContext.HelmServiceAPI.Upgrade(upgradeRequest, out)
	if err != nil {
		return "", err
	}
	return "", nil
}

func getHelmMessage(name string, args []string, environment *model.Environment, chart string) string {
	var message string

	message = "helm upgrade --install " + name + " \\\n"

	for _, e := range args {
		message = message + " --set \"" + e + "\" " + " \\\n"
	}
	message = message + " " + chart + " --namespace=" + environment.Namespace
	return message
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
