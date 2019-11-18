package dockerapi

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//DockerServiceBuilder DockerServiceBuilder
func DockerServiceBuilder() *DockerService {
	r := &DockerService{}
	r.httpClient = HTTPClientImpl{}
	return r
}

//HTTPClient HTTPClient
type HTTPClient interface {
	doRequest(url string, user string, password string) ([]byte, error)
}

//HTTPClientImpl HTTPClientImpl
type HTTPClientImpl struct {
	HTTPClient
}

//CacheInfo CacheInfo
type CacheInfo struct {
	imageName     string
	result        *model.ListDockerTagsResult
	repo          *model.DockerRepo
	matchFromDate bool
	dateFrom      time.Time
	globalCache   *sync.Map
}

func (h HTTPClientImpl) doRequest(url string, user string, password string) ([]byte, error) {

	client := getHTTPClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
	req.Header.Add("Authorization", " Basic "+sEnc)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	return bodyBytes, err

}

// DockerService is used to concretize DockerServiceInterface
type DockerService struct {
	httpClient HTTPClient
}

// DockerServiceInterface can be used to interact with a remote Docker repo
type DockerServiceInterface interface {
	GetDockerTagsWithDate(payload model.ListDockerTagsRequest, dao repository.DockerDAOInterface, globalCache *sync.Map) (*model.ListDockerTagsResult, error)
	GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error)
	GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error)
}

func getImageWithoutRepo(image string) string {
	result := ""
	index := strings.Index(image, "/")
	result = image[index+1:]
	return result
}

func getHTTPClient() *http.Client {
	tr := &http.Transport{}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}

func getBaseDomainFromRepo(dao repository.DockerDAOInterface, imageName string) (*model.DockerRepo, error) {
	firstBarIndex := strings.Index(imageName, "/")
	if firstBarIndex <= 0 {
		return nil, errors.New("Repository expected")
	}
	host := imageName[0:firstBarIndex]
	result, err := dao.GetDockerRepositoryByHost(host)
	return &result, err
}

func defineTagResponseFromCache(img string, tag string, createDate interface{}, cacheInfo CacheInfo) {

	if cacheInfo.matchFromDate {
		var object interface{}
		var dateTime time.Time
		object, ok := cacheInfo.globalCache.Load(img)
		if ok && object != nil {
			dateTime = object.(time.Time)
			major := dateTime.After(cacheInfo.dateFrom)
			if major {
				cacheInfo.result.TagResponse = append(cacheInfo.result.TagResponse, model.TagResponse{Tag: tag, Created: createDate.(time.Time)})
			}
		}

	} else {
		cacheInfo.result.TagResponse = append(cacheInfo.result.TagResponse, model.TagResponse{Tag: tag, Created: createDate.(time.Time)})
	}
}

func (docker DockerService) defineTagResponse(img string, tag string, cacheInfo CacheInfo) error {

	date, err := docker.GetDate(*cacheInfo.repo, cacheInfo.imageName, tag)
	if err != nil {
		return err
	}
	if cacheInfo.matchFromDate {
		if date.After(cacheInfo.dateFrom) {
			cacheInfo.result.TagResponse = append(cacheInfo.result.TagResponse, model.TagResponse{Tag: tag, Created: *date})
		}
	} else {
		cacheInfo.result.TagResponse = append(cacheInfo.result.TagResponse, model.TagResponse{Tag: tag, Created: *date})
	}
	cacheInfo.globalCache.Store(img, *date)
	return nil
}

func (docker DockerService) cacheDockerTags(tags []string, cacheInfo CacheInfo) error {
	for _, tag := range tags {
		img := cacheInfo.imageName + ":" + tag
		createDate, ok := cacheInfo.globalCache.Load(img)
		if ok {
			defineTagResponseFromCache(img, tag, createDate, cacheInfo)
		} else {
			err := docker.defineTagResponse(img, tag, cacheInfo)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//GetDockerTagsWithDate Method
func (docker DockerService) GetDockerTagsWithDate(payload model.ListDockerTagsRequest, dao repository.DockerDAOInterface, globalCache *sync.Map) (*model.ListDockerTagsResult, error) {

	var dateFrom time.Time
	matchFromDate := false
	if payload.From != "" {
		layout := "2006-01-02"
		dateFrom, _ = time.Parse(layout, payload.From)
		matchFromDate = true
	}

	repo, err := getBaseDomainFromRepo(dao, payload.ImageName)
	if err != nil {
		return nil, err
	}

	tagResult, err := docker.GetTags(repo, payload.ImageName)
	if err != nil {
		return nil, err
	}

	result := &model.ListDockerTagsResult{}

	cacheInfo := CacheInfo{}
	cacheInfo.imageName = payload.ImageName
	cacheInfo.result = result
	cacheInfo.repo = repo
	cacheInfo.matchFromDate = matchFromDate
	cacheInfo.dateFrom = dateFrom
	cacheInfo.globalCache = globalCache

	cacheErr := docker.cacheDockerTags(tagResult.Tags, cacheInfo)
	if cacheErr != nil {
		return nil, cacheErr
	}

	sort.Slice(result.TagResponse, func(i, j int) bool {
		return result.TagResponse[i].Created.Before(result.TagResponse[j].Created)
	})

	return result, nil

}

// GetDate fetches the docker image creation from a remote repo
func (docker DockerService) GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error) {

	imageName = getImageWithoutRepo(imageName)

	url := "https://" + repo.Host + "/v2/" + imageName + "/manifests/" + tag

	body, err := docker.httpClient.doRequest(url, repo.Username, repo.Password)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(bytes.NewReader(body))
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

	if len(list) <= 0 {
		currentDate := time.Now()
		return &currentDate, nil
	}

	return &list[len(list)-1].Created, nil

}

// GetTags fetches the docker tags from a remote repo
func (docker DockerService) GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error) {

	url := "https://" + repo.Host + "/v2/" + getImageWithoutRepo(imageName) + "/tags/list"

	body, err := docker.httpClient.doRequest(url, repo.Username, repo.Password)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(bytes.NewReader(body))
	var tagResult model.TagsResult

	err = d.Decode(&tagResult)
	if err != nil {
		return nil, err
	}
	return &tagResult, nil

}
