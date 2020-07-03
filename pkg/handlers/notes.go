package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/softplan/tenkai-api/pkg/global"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) newNotes(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.Notes

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := appContext.Repositories.NotesDAO.CreateNotes(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) editNotes(w http.ResponseWriter, r *http.Request) {

	var payload model.Notes

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.NotesDAO.EditNotes(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) findNotesByServiceName(w http.ResponseWriter, r *http.Request) {

	serviceNames, ok := r.URL.Query()["serviceName"]

	if !ok || len(serviceNames[0]) < 1 {
		log.Println("Url Param 'serviceName' is missing")
		return
	}

	serviceName := serviceNames[0]

	var notes *model.Notes
	var err error
	if notes, err = appContext.Repositories.NotesDAO.GetByServiceName(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(notes)
	w.Write(data)

}
