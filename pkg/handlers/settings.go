package handlers

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
)

func (appContext *AppContext) addSettings(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.SettingsList

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, e := range payload.List {
		configMap := model.ConfigMap{Name: e.Name, Value: e.Value}
		if _, err := appContext.Repositories.ConfigDAO.CreateOrUpdateConfig(configMap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) getSettingList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetSettingsListRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var settings model.SettingsList

	for _, name := range payload.List {

		var config model.ConfigMap
		var err error
		if config, err = appContext.Repositories.ConfigDAO.GetConfigByName(name); err != nil {
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
