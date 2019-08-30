package main

import (
	"bytes"
	"errors"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type releaseToDeploy struct {
	Name string
	Chart string
}

func (appContext *appContext) promote(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	if !contains(principal.Roles, TenkaiAdmin) {
		http.Error(w, errors.New("Acccess Denied!").Error(), http.StatusUnauthorized)
		return
	}

	out := &bytes.Buffer{}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	srcEnvID, ok := r.URL.Query()["srcEnvID"]
	if !ok || len(srcEnvID[0]) < 1 {
		http.Error(w, errors.New("param srcEnvID is required").Error(), 501)
		return
	}

	targetEnvID, ok := r.URL.Query()["targetEnvID"]
	if !ok || len(srcEnvID[0]) < 1 {
		http.Error(w, errors.New("param targetEnvID is required").Error(), 501)
		return
	}

	srcEnvIDi, _ := strconv.ParseInt(srcEnvID[0], 10, 64)
	targetEnvIDi, _ := strconv.ParseInt(targetEnvID[0], 10, 64)


	srcEnvironment, err := appContext.database.GetByID(int(srcEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	targetEnvironment, err := appContext.database.GetByID(int(targetEnvIDi))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	has, err := appContext.hasAccess(principal.Email, int(srcEnvironment.ID))
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in environment" + srcEnvironment.Namespace).Error(), http.StatusUnauthorized)
		return
	}

	has, err = appContext.hasAccess(principal.Email, int(targetEnvironment.ID))
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in environment" + targetEnvironment.Namespace).Error(), http.StatusUnauthorized)
		return
	}

	kubeConfig := global.KubeConfigBasePath + srcEnvironment.Group + "_" + srcEnvironment.Name

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

	toDeploy, err := retrieveReleasesToDeploy(&appContext.mutex, kubeConfig, srcEnvironment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	toPurge, err := retrieveReleasesToPurge(&appContext.mutex, kubeConfig, targetEnvironment.Namespace)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	err = appContext.purgeAll(kubeConfig, toPurge)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	for _, e := range toDeploy {
		appContext.simpleInstall(int(targetEnvironment.ID), e.Chart, e.Name, out)
	}

	w.WriteHeader(http.StatusOK)

}


func (appContext *appContext) purgeAll(kubeConfig string, envs []releaseToDeploy ) error {
	for _, e := range envs {
		err := helmapi.DeleteHelmRelease(kubeConfig, e.Name, true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (appContext *appContext) deleteEnvironmentVariables(envID uint) error {
	err := appContext.database.DeleteVariableByEnvironmentID(int(envID))
	return err
}

func (appContext *appContext) copyEnvironmentVariablesFromSrcToTarget(srcEnvID uint, targetEnvID uint) error {

	variables, err := appContext.database.GetAllVariablesByEnvironment(int(srcEnvID))
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

		if err := appContext.database.CreateVariable(*newVariable); err != nil {
			return err
		}
	}

	return nil

}


func retrieveReleasesToDeploy(mutex *sync.Mutex, kubeConfig string, srcNamespace string) ([]releaseToDeploy, error) {
	result := make([]releaseToDeploy, 0)
	mutex.Lock()
	list, err := helmapi.ListHelmDeployments(kubeConfig, srcNamespace)
	mutex.Unlock()
	if err != nil {
		return result, err
	}
	for _, e := range list.Releases {
		name := strings.ReplaceAll(e.Name, "-" + srcNamespace, "")
		lastHifen := strings.LastIndex(e.Chart, "-")

		//TODO - FIND A WAY TO DEFINE THE RIGHT REPOSITORY
		chart := "saj6/" + e.Chart[:lastHifen]
		result = append(result, releaseToDeploy{Name: name,Chart:chart})
	}
	return result, nil
}

func retrieveReleasesToPurge(mutex *sync.Mutex, kubeConfig string, namespace string) ([]releaseToDeploy, error) {
	result := make([]releaseToDeploy, 0)
	mutex.Lock()
	list, err := helmapi.ListHelmDeployments(kubeConfig, namespace)
	mutex.Unlock()
	if err != nil {
		return result, err
	}
	for _, e := range list.Releases {

		lastHifen := strings.LastIndex(e.Chart, "-")

		//TODO - FIND A WAY TO DEFINE THE RIGHT REPOSITORY
		chart := "saj6/" + e.Chart[:lastHifen]
		result = append(result, releaseToDeploy{Name: e.Name,Chart:chart})
	}
	return result, nil
}