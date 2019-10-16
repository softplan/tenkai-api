package handlers

import (
	"bytes"
	"errors"
	audit2 "github.com/softplan/tenkai-api/pkg/audit"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/helm"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
	"strings"
)

type releaseToDeploy struct {
	Name         string
	Chart        string
	ChartVersion string
}

func (appContext *AppContext) promote(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	principal := util.GetPrincipal(r)

	if !util.Contains(principal.Roles, constraints.TenkaiPromote) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
		return
	}

	modes, ok := r.URL.Query()["mode"]

	if !ok || len(modes[0]) < 1 {
		http.Error(w, errors.New("Url Param 'mode' is missing").Error(), http.StatusUnauthorized)
		return
	}

	mode := modes[0]

	srcEnvID, ok := r.URL.Query()["srcEnvID"]
	if !ok || len(srcEnvID[0]) < 1 || srcEnvID[0] == "undefined" {
		http.Error(w, errors.New("param srcEnvID is required").Error(), 501)
		return
	}

	targetEnvID, ok := r.URL.Query()["targetEnvID"]
	if !ok || len(targetEnvID[0]) < 1 || targetEnvID[0] == "undefined" {
		http.Error(w, errors.New("param targetEnvID is required").Error(), 501)
		return
	}

	srcEnvIDi, _ := strconv.ParseInt(srcEnvID[0], 10, 64)
	targetEnvIDi, _ := strconv.ParseInt(targetEnvID[0], 10, 64)

	srcEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(srcEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	targetEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(targetEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	has, err := appContext.hasAccess(principal.Email, int(srcEnvironment.ID))
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in environment "+srcEnvironment.Namespace).Error(), http.StatusUnauthorized)
		return
	}

	has, err = appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in environment "+targetEnvironment.Namespace).Error(), http.StatusUnauthorized)
		return
	}

	kubeConfig := global.KubeConfigBasePath + srcEnvironment.Group + "_" + srcEnvironment.Name

	if mode == "full" {

		err = appContext.deleteEnvironmentVariables(targetEnvironment.ID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		err = appContext.copyEnvironmentVariablesFromSrcToTarget(srcEnvironment.ID, targetEnvironment.ID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

	} else {

		err = appContext.copyImageAndTagFromSrcToTarget(srcEnvironment.ID, targetEnvironment.ID)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}

	}

	toPurge, err := retrieveReleasesToPurge(appContext.HelmServiceAPI, kubeConfig, targetEnvironment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	toDeploy, err := retrieveReleasesToDeploy(appContext.HelmServiceAPI, kubeConfig, srcEnvironment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	go appContext.doIt(kubeConfig, targetEnvironment, toPurge, toDeploy)

	auditValues := make(map[string]string)
	auditValues["sourceEnvironment"] = srcEnvironment.Name
	auditValues["targetEnvironment"] = targetEnvironment.Name
	auditValues["mode"] = mode

	audit2.DoAudit(r.Context(), appContext.Elk, principal.Email, "promote", auditValues)

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) doIt(kubeConfig string, targetEnvironment *model.Environment, toPurge []releaseToDeploy, toDeploy []releaseToDeploy) {

	out := &bytes.Buffer{}

	logFields := global.AppFields{global.Function: "doIt - promoting", "target": targetEnvironment.Name}

	err := appContext.purgeAll(kubeConfig, toPurge)
	if err != nil {
		global.Logger.Error(logFields, "error: "+err.Error())
		return
	}

	for _, e := range toDeploy {
		global.Logger.Info(logFields, "deploying: "+e.Name+" - "+e.Chart)
		_, err := appContext.simpleInstall(targetEnvironment, e.Chart, e.ChartVersion, e.Name, out, false, false)
		if err != nil {
			global.Logger.Error(logFields, "error: "+err.Error())
		}
	}

}

func (appContext *AppContext) purgeAll(kubeConfig string, envs []releaseToDeploy) error {
	for _, e := range envs {
		err := appContext.HelmServiceAPI.DeleteHelmRelease(kubeConfig, e.Name, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (appContext *AppContext) deleteEnvironmentVariables(envID uint) error {
	err := appContext.Repositories.VariableDAO.DeleteVariableByEnvironmentID(int(envID))
	return err
}

func (appContext *AppContext) copyEnvironmentVariablesFromSrcToTarget(srcEnvID uint, targetEnvID uint) error {

	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(int(srcEnvID))
	if err != nil {
		return err
	}

	var newVariable *model.Variable
	for _, variable := range variables {
		newVariable = &model.Variable{}
		newVariable.Name = variable.Name
		newVariable.EnvironmentID = int(targetEnvID)
		newVariable.Value = variable.Value
		newVariable.Description = variable.Description
		newVariable.Scope = variable.Scope

		if _, _, err := appContext.Repositories.VariableDAO.CreateVariable(*newVariable); err != nil {
			return err
		}
	}

	return nil

}

func (appContext *AppContext) copyImageAndTagFromSrcToTarget(srcEnvID uint, targetEnvID uint) error {

	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(int(srcEnvID))
	if err != nil {
		return err
	}

	var newVariable *model.Variable
	for _, variable := range variables {

		if variable.Name == "image.tag" || variable.Name == "image.repository" {

			newVariable = &model.Variable{}
			newVariable.Name = variable.Name
			newVariable.EnvironmentID = int(targetEnvID)
			newVariable.Value = variable.Value
			newVariable.Description = variable.Description
			newVariable.Scope = variable.Scope

			if _, _, err := appContext.Repositories.VariableDAO.CreateVariable(*newVariable); err != nil {
				return err
			}

		}
	}

	return nil

}

func retrieveReleasesToDeploy(hsi helmapi.HelmServiceInterface, kubeConfig string, srcNamespace string) ([]releaseToDeploy, error) {
	result := make([]releaseToDeploy, 0)
	list, err := hsi.ListHelmDeployments(kubeConfig, srcNamespace)
	if err != nil {
		return result, err
	}
	for _, e := range list.Releases {
		name := strings.ReplaceAll(e.Name, "-"+srcNamespace, "")
		lastHifen := strings.LastIndex(e.Chart, "-")

		//TODO - FIND A WAY TO DEFINE THE RIGHT REPOSITORY
		chart := "saj6/" + e.Chart[:lastHifen]
		result = append(result, releaseToDeploy{Name: name, Chart: chart})
	}
	return result, nil
}

func retrieveReleasesToPurge(hsi helmapi.HelmServiceInterface, kubeConfig string, namespace string) ([]releaseToDeploy, error) {

	result := make([]releaseToDeploy, 0)
	list, err := hsi.ListHelmDeployments(kubeConfig, namespace)

	if err != nil {
		return result, err
	}

	if list == nil {
		return result, nil
	}

	for _, e := range list.Releases {

		lastHifen := strings.LastIndex(e.Chart, "-")

		//TODO - FIND A WAY TO DEFINE THE RIGHT REPOSITORY
		chart := "saj6/" + e.Chart[:lastHifen]
		result = append(result, releaseToDeploy{Name: e.Name, Chart: chart})
	}
	return result, nil
}
