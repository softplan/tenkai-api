package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"github.com/streadway/amqp"
)

const defaultRepo = "DEFAULT_REPO_"

func (appContext *AppContext) publishRepoMessage(exchange string, body interface{}) error {
	bodyBytes, _ := json.Marshal(body)

	if err := appContext.RabbitImpl.Publish(
		appContext.RabbitMQChannel,
		exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        bodyBytes,
		},
	); err != nil {
		global.Logger.Error(global.AppFields{global.Function: "repoUpdate"}, "Error on publish a message "+err.Error())
		return err
	}
	return nil
}

func (appContext *AppContext) repoUpdate(w http.ResponseWriter, r *http.Request) {
	if err := appContext.publishRepoMessage("update.repository.fx", nil); err != nil {
		http.Error(w, "Error publishing repoUpdate on queue", http.StatusInternalServerError)
	}
	appContext.HelmServiceAPI.RepoUpdate()
}

func (appContext *AppContext) listRepositories(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	result := &model.RepositoryResult{}

	repositories, err := appContext.HelmServiceAPI.GetRepositories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result.Repositories = repositories
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) newRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
	}

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.Repository

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.HelmServiceAPI.AddRepository(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.publishRepoMessage("add.repository.fx", payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) setDefaultRepo(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.DefaultRepoRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	configMap := model.ConfigMap{Name: defaultRepo + principal.Email, Value: payload.Reponame}

	if _, err := appContext.Repositories.ConfigDAO.CreateOrUpdateConfig(configMap); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) getDefaultRepo(w http.ResponseWriter, r *http.Request) {
	principal := util.GetPrincipal(r)

	config, err := appContext.Repositories.ConfigDAO.GetConfigByName(defaultRepo + principal.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(config)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deleteRepository(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	name := vars["name"]
	w.Header().Set(global.ContentType, global.JSONContentType)
	if err := appContext.HelmServiceAPI.RemoveRepository(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.publishRepoMessage("del.repository.fx", name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) getModelRepositoryDefault(principal model.Principal) (model.Repository, error) {
	var repo model.Repository

	config, err := appContext.Repositories.ConfigDAO.GetConfigByName(defaultRepo + principal.Email)
	if err != nil {
		return repo, err
	}
	repositories, err := appContext.HelmServiceAPI.GetRepositories()
	if err != nil {
		return repo, err
	}
	for _, repository := range repositories {
		if config.Value == repository.Name {
			return repository, nil
		}
	}

	return repo, nil
}
