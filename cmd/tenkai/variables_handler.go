package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
)

func (appContext *appContext) deleteVariable(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiVariablesDelete) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.database.DeleteVariable(id); err != nil {
		log.Println("Error deleting variable: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) editVariable(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiVariablesSave) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	var payload model.DataVariableElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payload.Data.Secret {
		secret := util.Encrypt([]byte(payload.Data.Value), appContext.configuration.App.Passkey)
		payload.Data.Value = hex.EncodeToString(secret)
	}

	if err := appContext.database.EditVariable(payload.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) addVariables(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !contains(principal.Roles, TenkaiVariablesSave) {
		http.Error(w, errors.New("Acccess Defined").Error(), http.StatusUnauthorized)
	}

	var payload model.DataVariableElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payload.Data.Secret {
		secret := util.Encrypt([]byte(payload.Data.Value), appContext.configuration.App.Passkey)
		payload.Data.Value = hex.EncodeToString(secret)
	}

	if err := appContext.database.EditVariable(payload.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) getVariables(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	vars := mux.Vars(r)
	sl := vars["envId"]
	id, _ := strconv.Atoi(sl)
	variableResult := &model.VariablesResult{}

	has, failed := appContext.hasAccess(principal.Email, id)
	if failed != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	var err error
	if variableResult.Variables, err = appContext.database.GetAllVariablesByEnvironment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, e := range variableResult.Variables {
		if e.Secret {
			byteValues, _ := hex.DecodeString(e.Value)
			value := util.Decrypt(byteValues, appContext.configuration.App.Passkey)
			variableResult.Variables[i].Value = string(value)

		}
	}

	data, _ := json.Marshal(variableResult)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}
