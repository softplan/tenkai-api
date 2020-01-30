package handlers

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

func (appContext *AppContext) deleteVariable(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiVariablesDelete) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set(global.ContentType, global.JSONContentType)
	if err := appContext.Repositories.VariableDAO.DeleteVariable(id); err != nil {
		log.Println("Error deleting variable: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) editVariable(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiVariablesSave) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	var payload model.DataVariableElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if payload.Data.Secret {
		secret := util.Encrypt([]byte(payload.Data.Value), appContext.Configuration.App.Passkey)
		payload.Data.Value = hex.EncodeToString(secret)
	}

	if err := appContext.Repositories.VariableDAO.EditVariable(payload.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) getVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

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
	if variableResult.Variables, err = appContext.Repositories.VariableDAO.GetAllVariablesByEnvironment(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appContext.decodeSecrets(variableResult)

	data, _ := json.Marshal(variableResult)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) copyVariableValue(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiVariablesSave) {
		http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
		return
	}

	var payload model.CopyVariableValue
	var err error

	if err = util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var sourceVar *model.Variable
	if sourceVar, err = appContext.Repositories.VariableDAO.GetByID(payload.SrcVarID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetVar *model.Variable
	if targetVar, err = appContext.Repositories.VariableDAO.GetByID(payload.TarVarID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	targetVar.Value = sourceVar.Value
	if err := appContext.Repositories.VariableDAO.EditVariable(*targetVar); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}
