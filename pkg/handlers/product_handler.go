package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	analyser "github.com/softplan/tenkai-api/pkg/service/analyser"
	"github.com/softplan/tenkai-api/pkg/util"
)

const pvLockMsg = "Product version locked"

func (appContext *AppContext) newProduct(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.Product

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := appContext.Repositories.ProductDAO.CreateProduct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) editProduct(w http.ResponseWriter, r *http.Request) {

	var payload model.Product

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.ProductDAO.EditProduct(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) deleteProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set(global.ContentType, global.JSONContentType)
	if err := appContext.Repositories.ProductDAO.DeleteProduct(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) listProducts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	result := &model.ProductRequestReponse{}
	var err error
	if result.List, err = appContext.Repositories.ProductDAO.ListProducts(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) newProductVersion(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.ProductVersion

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	payload.Date = time.Now()

	if _, err := appContext.Repositories.ProductDAO.CreateProductVersionCopying(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) deleteProductVersion(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set(global.ContentType, global.JSONContentType)

	// Deletes ProductVersionServices
	childs := &model.ProductVersionServiceRequestReponse{}
	var err error
	if childs.List, err = appContext.Repositories.ProductDAO.ListProductsVersionServices(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, e := range childs.List {
		if err := appContext.Repositories.ProductDAO.DeleteProductVersionService(int(e.ID)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Deletes ProductVersion itself
	if err = appContext.Repositories.ProductDAO.DeleteProductVersion(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) newProductVersionService(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.ProductVersionService

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pv, err := appContext.Repositories.ProductDAO.ListProductVersionsByID(payload.ProductVersionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if pv.Locked {
		http.Error(w, pvLockMsg, http.StatusInternalServerError)
		return
	}

	if !appContext.validateVersion(pv.Version, payload.DockerImageTag) {
		http.Error(w, "Wrong version", http.StatusInternalServerError)
		return
	}

	if _, err := appContext.Repositories.ProductDAO.CreateProductVersionService(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) editProductVersionService(w http.ResponseWriter, r *http.Request) {
	var payload model.ProductVersionService

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pv, err := appContext.Repositories.ProductDAO.ListProductVersionsByID(payload.ProductVersionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.ProductDAO.EditProductVersionService(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if pv.Locked {
		http.Error(w, pvLockMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) deleteProductVersionService(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	sl := vars["id"]
	id, _ := strconv.Atoi(sl)
	w.Header().Set(global.ContentType, global.JSONContentType)

	pvs, err := appContext.Repositories.ProductDAO.ListProductVersionsServiceByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pv, err := appContext.Repositories.ProductDAO.ListProductVersionsByID(pvs.ProductVersionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if pv.Locked {
		http.Error(w, pvLockMsg, http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.ProductDAO.DeleteProductVersionService(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) listProductVersions(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	ids, ok := r.URL.Query()["productId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.ProductVersionRequestReponse{}
	var err error

	if result.List, err = appContext.Repositories.ProductDAO.ListProductsVersions(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) lockUnlockCommon(w http.ResponseWriter, r *http.Request) (*model.ProductVersion, int, error) {
	principal := util.GetPrincipal(r)
	if !util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		return nil, http.StatusUnauthorized, errors.New(global.AccessDenied)
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	pv, e := appContext.Repositories.ProductDAO.ListProductVersionsByID(id)
	if e != nil {
		return nil, http.StatusInternalServerError, e
	}

	return pv, http.StatusOK, nil
}

func (appContext *AppContext) lockProductVersion(w http.ResponseWriter, r *http.Request) {
	pv, httpCode, err := appContext.lockUnlockCommon(w, r)
	if err != nil {
		http.Error(w, err.Error(), httpCode)
		return
	}

	pv.Locked = true

	if err := appContext.Repositories.ProductDAO.EditProductVersion(*pv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) unlockProductVersion(w http.ResponseWriter, r *http.Request) {
	pv, httpCode, err := appContext.lockUnlockCommon(w, r)
	if err != nil {
		http.Error(w, err.Error(), httpCode)
		return
	}

	pv.Locked = false

	if err := appContext.Repositories.ProductDAO.EditProductVersion(*pv); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) listProductVersionServices(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	ids, ok := r.URL.Query()["productVersionId"]
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, _ := strconv.Atoi(ids[0])
	result := &model.ProductVersionServiceRequestReponse{}
	var err error

	if result.List, err = appContext.Repositories.ProductDAO.ListProductsVersionServices(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wg := new(sync.WaitGroup)

	helmCharts := make(map[string][]model.SearchResult)

	for i, e := range result.List {

		if e.ServiceName != "" && e.DockerImageTag != "" {

			var serviceName = e.ServiceName
			var tag = e.DockerImageTag
			var helmRepo = splitChartRepo(serviceName)
			index := i

			if helmCharts[helmRepo] == nil {
				helmCharts[helmRepo] = *appContext.HelmServiceAPI.SearchCharts([]string{helmRepo}, false)
			}

			wg.Add(1)
			go func(wg *sync.WaitGroup, serviceName string, tag string, index int, searchResult []model.SearchResult) {
				defer wg.Done()
				version, _ := appContext.verifyNewVersion(splitSrvNameIfNeeded(serviceName), tag)
				result.List[index].LatestVersion = version
				result.List[index].ChartLatestVersion = appContext.getChartLatestVersion(serviceName, searchResult)
			}(wg, serviceName, tag, index, helmCharts[helmRepo])
		}
	}

	wg.Wait()

	sort.Slice(result.List, func(i, j int) bool {
		return result.List[i].ServiceName > (result.List[j].ServiceName)
	})

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

//splitSrvNameIfNeeded Returns service name removing chart version
func splitSrvNameIfNeeded(serviceName string) string {
	svcName := strings.Split(serviceName, " - ")
	if len(svcName) == 2 {
		serviceName = svcName[0]
	}
	return serviceName
}

//splitChartVersion Returns chart version removing service name
func splitChartVersion(serviceName string) string {
	splited := strings.Split(serviceName, " - ")
	if len(splited) == 2 {
		return splited[1]
	}
	return ""
}

//splitChartRepo Return chart repo removing service name
func splitChartRepo(serviceName string) string {
	repo := strings.Split(serviceName, "/")
	if len(repo) == 2 {
		return repo[0]
	}
	return ""
}

func (appContext *AppContext) getChartLatestVersion(serviceName string, charts []model.SearchResult) string {
	var currentChartVersion = splitChartVersion(serviceName)
	serviceName = splitSrvNameIfNeeded(serviceName)
	for _, sr := range charts {
		if sr.Name == serviceName && sr.ChartVersion != currentChartVersion {
			return sr.ChartVersion
		}
	}

	return ""
}

func getCreateDateOfCurrentTag(tags []model.TagResponse, dockerImageTag string) time.Time {
	var currentDate time.Time
	for _, e := range tags {
		if e.Tag == dockerImageTag {
			currentDate = e.Created
			break
		}
	}
	return currentDate
}

func (appContext *AppContext) getImageName(serviceName string) (string, error) {

	var result string

	object, ok := appContext.ChartImageCache.Load(serviceName)
	var imageCache string
	if ok {
		imageCache = object.(string)
	}

	if !ok || imageCache == "" {
		var err error

		result, err = analyser.GetImageFromService(appContext.HelmServiceAPI, serviceName, &appContext.Mutex)
		if err != nil {
			return "", err
		}

		appContext.ChartImageCache.Store(serviceName, result)

	} else {
		object, ok := appContext.ChartImageCache.Load(serviceName)
		if ok {
			result = object.(string)
		}
	}
	return result, nil
}

func (appContext *AppContext) getDockerTagsWithDate(p model.ListDockerTagsRequest) (model.ListDockerTagsResult, error) {
	var result model.ListDockerTagsResult

	url := fmt.Sprintf("%s/listDockerTags", appContext.Configuration.App.DockerAPIURL)
	m, _ := json.Marshal(p)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(m))
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func (appContext *AppContext) verifyNewVersion(serviceName string, dockerImageTag string) (string, error) {

	currentTag := getNumberOfTag(dockerImageTag)

	var payload model.ListDockerTagsRequest
	var err error

	payload.ImageName, err = appContext.getImageName(serviceName)
	if err != nil {
		return "", err
	}

	//Get version tags
	result, err := appContext.getDockerTagsWithDate(payload)
	if err != nil {
		return "", err
	}

	var currentDate time.Time
	majorList := make([]model.TagResponse, 0)

	currentDate = getCreateDateOfCurrentTag(result.TagResponse, dockerImageTag)

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

func getNumberOfTag(tag string) uint64 {

	//Count amount of delimiters
	n := strings.Count(tag, "#")
	n = n + strings.Count(tag, ".")
	n = n + strings.Count(tag, "-")

	// Normalize delimiters to dot
	r := strings.ReplaceAll(tag, "#", ".")
	r = strings.ReplaceAll(r, ".", ".")
	r = strings.ReplaceAll(r, "-", ".")

	// Add leading zeros when length is one
	resultStr := ""
	for _, s := range strings.Split(r, ".") {
		if len(s) == 1 {
			s = "0" + s
		}
		resultStr += s
	}

	for i := 0; i < 6-n; i++ {
		resultStr = resultStr + "00"
	}

	result, _ := strconv.ParseUint(resultStr, 10, 64)

	return result
}

func (appContext *AppContext) validateVersion(productVersion string, currentVersion string) bool {
	a := strings.Split(normalize(productVersion), ".")
	b := strings.Split(normalize(currentVersion), ".")

	if len(a) >= 3 && len(b) >= 3 {
		return a[0] == b[0] && a[1] == b[1] && a[2] == b[2]
	}

	return false
}

func normalize(s string) string {
	return strings.ReplaceAll(s, "-", ".")
}
