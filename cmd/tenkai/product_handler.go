package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	analyser "github.com/softplan/tenkai-api/service/tenkai"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (appContext *appContext) newProduct(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.Product

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := appContext.database.CreateProduct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) editProduct(w http.ResponseWriter, r *http.Request) {

	var payload model.Product

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.database.EditProduct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) deleteProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := appContext.database.DeleteProduct(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *appContext) listProducts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	result := &model.ProductRequestReponse{}
	var err error
	if result.List, err = appContext.database.ListProducts(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) newProductVersion(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.ProductVersion

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload.Date = time.Now()
	if _, err := appContext.database.CreateProductVersion(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) newProductVersionService(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	var payload model.ProductVersionService

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := appContext.database.CreateProductVersionService(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *appContext) editProductVersionService(w http.ResponseWriter, r *http.Request) {
	var payload model.ProductVersionService

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.database.EditProductVersionService(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *appContext) listProductVersions(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	ids, ok := r.URL.Query()["productId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.ProductVersionRequestReponse{}
	var err error

	if result.List, err = appContext.database.ListProductsVersions(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) listProductVersionServices(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	ids, ok := r.URL.Query()["productVersionId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.ProductVersionServiceRequestReponse{}
	var err error

	if result.List, err = appContext.database.ListProductsVersionServices(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, e := range result.List {
		if e.ServiceName != "" && e.DockerImageTag != "" {
			_ = appContext.verifyNewVersion(&e)
			result.List[i].LatestVersion = e.LatestVersion
		}
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) verifyNewVersion(pvs *model.ProductVersionService) error {

	var payload model.ListDockerTagsRequest
	var err error
	payload.ImageName, err = analyser.GetImageFromService(pvs.ServiceName)
	if err != nil {
		return err
	}

	//Get version tags
	result, err := dockerapi.GetDockerTagsWithDate(payload, appContext.testMode, appContext.database, appContext.dockerTagsCache)
	if err != nil {
		return err
	}

	var currentDate time.Time
	majorList := make([]model.TagResponse, 0)

	//Get create date of current tag
	for _, e := range result.TagResponse {
		if e.Tag == pvs.DockerImageTag {
			currentDate = e.Created
			break
		}
	}

	//Get all tags created after current tag
	for _, e := range result.TagResponse {
		if e.Created.After(currentDate) {
			majorList = append(majorList, e)
		}
	}

	finalList := make([]model.TagResponse, 0)

	//Filter based on version tag
	for _, e := range majorList {
		if getNumberOfTag(e.Tag) > getNumberOfTag(pvs.DockerImageTag) {
			finalList = append(finalList, e)
		}
	}

	if len(finalList) > 0 {
		e := finalList[len(finalList)-1]
		pvs.LatestVersion = e.Tag
	}

	return nil

}

func getNumberOfTag(tag string) int {

	//Count amount of delimiters
	n := strings.Count(tag, "#")
	n = n + strings.Count(tag, ".")
	n = n + strings.Count(tag, "-")

	for i := 0; i < 10 - n; i++ {
		tag = tag + ".00"
	}

	tag = strings.ReplaceAll(tag, "#", "")
	tag = strings.ReplaceAll(tag, ".", "")
	tag = strings.ReplaceAll(tag, "-", "")
	result, _ := strconv.Atoi(tag)

	return result
}
