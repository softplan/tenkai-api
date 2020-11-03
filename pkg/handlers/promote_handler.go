package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/util"
)

type releaseToDeploy struct {
	Name         string
	Chart        string
	ChartVersion string
}

func (appContext *AppContext) validateAndExtractParams(w http.ResponseWriter, r *http.Request) (string, int64, int64, error) {

	modes, ok := r.URL.Query()["mode"]

	if !ok || len(modes[0]) < 1 {
		error := errors.New("Url Param 'mode' is missing")
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return "", -1, -1, error
	}

	mode := modes[0]

	srcEnvID, ok := r.URL.Query()["srcEnvID"]
	if !ok || len(srcEnvID[0]) < 1 || srcEnvID[0] == "undefined" {
		error := errors.New("param srcEnvID is required")
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return "", -1, -1, error
	}

	targetEnvID, ok := r.URL.Query()["targetEnvID"]
	if !ok || len(targetEnvID[0]) < 1 || targetEnvID[0] == "undefined" {
		error := errors.New("param targetEnvID is required")
		http.Error(w, error.Error(), http.StatusInternalServerError)
		return "", -1, -1, error
	}

	srcEnvIDi, _ := strconv.ParseInt(srcEnvID[0], 10, 64)
	targetEnvIDi, _ := strconv.ParseInt(targetEnvID[0], 10, 64)

	return mode, srcEnvIDi, targetEnvIDi, nil

}

func (appContext *AppContext) retrieveSrcAndTargetEnv(w http.ResponseWriter, principal model.Principal, srcEnvIDi int64, targetEnvIDi int64) (*model.Environment, *model.Environment, error) {
	srcEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(srcEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return nil, nil, err
	}

	targetEnvironment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(targetEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return nil, nil, err
	}

	has, err := appContext.hasAccess(principal.Email, int(srcEnvironment.ID))
	if err != nil || !has {
		newErr := errors.New("Access Denied in environment " + srcEnvironment.Namespace)
		http.Error(w, newErr.Error(), http.StatusUnauthorized)
		return nil, nil, newErr
	}

	has, err = appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
	if err != nil || !has {
		newErr := errors.New("Access Denied in environment " + targetEnvironment.Namespace)
		http.Error(w, newErr.Error(), http.StatusUnauthorized)
		return nil, nil, newErr
	}

	return srcEnvironment, targetEnvironment, nil

}

func (appContext *AppContext) promote(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	principal := util.GetPrincipal(r)

	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied").Error(), http.StatusUnauthorized)
		return
	}

	mode, srcEnvIDi, targetEnvIDi, err := appContext.validateAndExtractParams(w, r)
	if err != nil {
		return
	}

	srcEnvironment, targetEnvironment, envErr := appContext.retrieveSrcAndTargetEnv(w, principal, srcEnvIDi, targetEnvIDi)
	if envErr != nil {
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(srcEnvironment.Group, srcEnvironment.Name)

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

	appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "promote", auditValues)

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

		installPayload := convertPayload(e)

		_, err := appContext.simpleInstall(targetEnvironment, installPayload, out, false, false, "")
		if err != nil {
			global.Logger.Error(logFields, "error: "+err.Error())
		}
	}

}

func convertPayload(e releaseToDeploy) model.InstallPayload {
	p := model.InstallPayload{}
	p.Chart = e.Chart
	p.ChartVersion = e.ChartVersion
	p.Name = e.Name
	return p
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

		chart := e.Chart[:lastHifen]
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

		chart := e.Chart[:lastHifen]
		result = append(result, releaseToDeploy{Name: e.Name, Chart: chart})
	}
	return result, nil
}
