package main

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"sort"
	"strings"
	"time"
)

func getImageWithoutRepo(image string) string {
	result := ""
	index := strings.Index(image, "/")
	result = image[index+1:]
	return result
}

func getHttpClient() *http.Client {
	tr := &http.Transport{}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}


func (appContext *appContext) listDockerTags(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.ListDockerTagsRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	repo, err := getBaseDomainFromRepo(&appContext.database, payload.ImageName)
	url := "https://" + repo.Host + "/v2/" + getImageWithoutRepo(payload.ImageName) + "/tags/list"

	client := getHttpClient()



	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(repo.Username + ":" + repo.Password))

	req.Header.Add("Authorization", " Basic "+sEnc)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var tagResult model.TagsResult

	err = d.Decode(&tagResult)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result := &model.ListDockerTagsResult{}

	for _, e := range tagResult.Tags {
		date, err := getDate(*repo, getImageWithoutRepo(payload.ImageName), e)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
		result.TagResponse = append(result.TagResponse, model.TagResponse{Created: *date, Tag: e})

	}

	sort.Slice(result.TagResponse, func(i, j int) bool {
		return result.TagResponse[i].Created.Before(result.TagResponse[j].Created)
	})

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func getDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error) {

	url := "https://" + repo.Host + "/v2/" + imageName + "/manifests/" + tag

	client := getHttpClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(repo.Username + ":" + repo.Password))
	req.Header.Add("Authorization", " Basic "+sEnc)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var manifestResult model.ManifestResult

	err = d.Decode(&manifestResult)
	if err != nil {
		return nil, err
	}

	list := make([]model.V1Compatibility, 0)

	for _, e := range manifestResult.History {
		v1Compatibility := &model.V1Compatibility{}
		json.Unmarshal([]byte(e.V1Compatibility), &v1Compatibility)
		list = append(list, *v1Compatibility)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Created.Before(list[j].Created)
	})

	return &list[len(list)-1].Created, nil

}

func getBaseDomainFromRepo(dbms *dbms.Database, imageName string) (*model.DockerRepo, error) {
	firstBarIndex := strings.Index(imageName, "/")
	host := imageName[0:firstBarIndex]
	result, err := dbms.GetDockerRepositoryByHost(host)
	return &result, err
}
