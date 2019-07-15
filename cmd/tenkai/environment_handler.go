package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/global"
	"github.com/softplan/tenkai-api/util"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/softplan/tenkai-api/dbms/model"
)

func (appContext *appContext) deleteEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	log.Println("Deleting environment: ", vars["id"])

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Println("Error processing parameter id: ", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	env, error := appContext.database.GetByID(id)
	if error != nil {
		log.Println("Error retrieving environment by ID: ", error)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := appContext.database.DeleteEnvironment(*env); err != nil {
		log.Println("Error deleting environment: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := removeEnvironmentFile(env.Group+"_"+env.Name); err != nil {
		log.Println("Error deleting environment file: ", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) editEnvironment(w http.ResponseWriter, r *http.Request) {

	var payload model.DataElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	env := payload.Data

	result, error := appContext.database.GetByID(int(env.ID))
	if error != nil {
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	oldFile := result.Group+"_"+result.Name
	removeEnvironmentFile(oldFile)

	createEnvironmentFile(env.Name, env.Token, env.Group+"_"+env.Name,
		env.CACertificate, env.ClusterURI, env.Namespace)

	if err := appContext.database.EditEnvironment(env); err != nil {
		if err := json.NewEncoder(w).Encode(error); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) addEnvironments(w http.ResponseWriter, r *http.Request) {

	var payload model.DataElement

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	env := payload.Data

	createEnvironmentFile(env.Name, env.Token, env.Group+"_"+env.Name,
		env.CACertificate, env.ClusterURI, env.Namespace)

	if err := appContext.database.CreateEnvironment(env); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)


}

func (appContext *appContext) getEnvironments(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	envResult := &model.EnvResult{}

	var err error
	if envResult.Envs, err = appContext.database.GetAllEnvironments(); err != nil {
		if err := json.NewEncoder(w).Encode(err); err != nil {
			if err := json.NewEncoder(w).Encode(err); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(envResult)
	w.Write(data)

}

func removeEnvironmentFile(fileName string) error {
	log.Println("Removing file: " + fileName)
	
	if _, err := os.Stat("./" + fileName); err == nil {
		err := os.Remove("./" + fileName)
		if err != nil {
			log.Println("Error removing file", err)
			return err
		}
	}
	return nil
}

func createEnvironmentFile(clusterName string, clusterUserToken string,
	fileName string, ca string, server string, namespace string) {

	file, err := os.Create(global.KUBECONFIG_BASE_PATH + fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	ca = strings.TrimSuffix(ca, "\n")
	caBase64 := base64.StdEncoding.EncodeToString([]byte(ca))

	startIndex := strings.Index(clusterUserToken, "kubeconfig-") + 11
	endIndex := strings.Index(clusterUserToken, ":")

	clusterUser := clusterUserToken[startIndex:endIndex]

	file.WriteString("apiVersion: v1\n")
	file.WriteString("clusters:\n")
	file.WriteString("- cluster:\n")
	file.WriteString("    certificate-authority-data: " + caBase64 + "\n")
	file.WriteString("    server: " + server + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("contexts:\n")
	file.WriteString("- context:\n")
	file.WriteString("    cluster: " + clusterName + "\n")
	file.WriteString("    namespace: " + namespace + "\n")
	file.WriteString("    user: " + clusterUser + "\n")
	file.WriteString("  name: " + clusterName + "\n")
	file.WriteString("current-context: " + clusterName + "\n")
	file.WriteString("kind: Config\n")
	file.WriteString("preferences: {}\n")
	file.WriteString("users:\n")
	file.WriteString("- name: " + clusterUser + "\n")
	file.WriteString("  user:\n")
	file.WriteString("    token: " + clusterUserToken + "\n")

}
