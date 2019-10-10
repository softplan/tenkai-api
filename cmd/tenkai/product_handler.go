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
	"sync"
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

	wg := new(sync.WaitGroup)

	for i, e := range result.List {

		if e.ServiceName != "" && e.DockerImageTag != "" {

			var serviceName = e.ServiceName
			var tag = e.DockerImageTag
			index := i

			wg.Add(1)
			go func(wg *sync.WaitGroup, serviceName string, tag string, index int) {
				defer wg.Done()
				version, _ := appContext.verifyNewVersion(serviceName, tag)
				result.List[index ].LatestVersion = version

			}(wg, serviceName, tag, index)
		}
	}

	wg.Wait()


	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *appContext) verifyNewVersion(serviceName string, dockerImageTag string) (string, error) {

	currentTag := getNumberOfTag(dockerImageTag)

	var payload model.ListDockerTagsRequest

	//imageCache := appContext.chartImageCache[pvs.ServiceName]
	object, ok := appContext.chartImageCache.Load(serviceName)
	var imageCache string
	if ok {
		imageCache = object.(string)
	}

	if !ok || imageCache == "" {
		var err error

		payload.ImageName, err = analyser.GetImageFromService(serviceName)
		if err != nil {
			return "", err
		}

		appContext.chartImageCache.Store(serviceName, payload.ImageName)

		//appContext.chartImageCache[pvs.ServiceName] = payload.ImageName


	} else {
		//payload.ImageName = appContext.chartImageCache[pvs.ServiceName]
		object, ok := appContext.chartImageCache.Load(serviceName)
		if ok {
			payload.ImageName = object.(string)
		}
	}

	//Get version tags
	result, err := dockerapi.GetDockerTagsWithDate(payload, appContext.testMode, appContext.database, appContext.dockerTagsCache)
	if err != nil {
		return "",err
	}

	var currentDate time.Time
	majorList := make([]model.TagResponse, 0)

	//Get create date of current tag
	for _, e := range result.TagResponse {
		if e.Tag == dockerImageTag {
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

		elementTag := getNumberOfTag(e.Tag)
		if elementTag > currentTag {
			finalList = append(finalList, e)
		}
	}

	var lastResult string
	if len(finalList) > 0 {
		e := finalList[len(finalList)-1]
		lastResult = e.Tag
	}

	return lastResult, nil

}

func getNumberOfTag(tag string) int {

	//Count amount of delimiters
	n := strings.Count(tag, "#")
	n = n + strings.Count(tag, ".")
	n = n + strings.Count(tag, "-")

	for i := 0; i < 10-n; i++ {
		tag = tag + ".00"
	}

	tag = strings.ReplaceAll(tag, "#", "")
	tag = strings.ReplaceAll(tag, ".", "")
	tag = strings.ReplaceAll(tag, "-", "")
	result, _ := strconv.Atoi(tag)

	return result
}
