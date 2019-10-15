package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"net/http"
)

func (appContext *appContext) addSettings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.SettingsList

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, e := range payload.List {
		configMap := model.ConfigMap{Name: e.Name, Value: e.Value}
		if _, err := appContext.repositories.configDAO.CreateOrUpdateConfig(configMap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) getSettingList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.GetSettingsListRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var settings model.SettingsList

	for _, name := range payload.List {

		var config model.ConfigMap
		var err error
		if config, err = appContext.repositories.configDAO.GetConfigByName(name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return

		}
		setting := model.Settings{Name: config.Name, Value: config.Value}
		settings.List = append(settings.List, setting)
	}

	data, _ := json.Marshal(settings)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
