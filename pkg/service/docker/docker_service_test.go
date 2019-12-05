package dockerapi

import (
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

//HTTPClientMock HTTPClientMock
type HTTPClientMock struct {
}

func (h HTTPClientMock) doRequest(url string, user string, password string) ([]byte, error) {
	return []byte("{\"name\":\"alfa\", \"tags\":[\"a\", \"b\"]}"), nil
}

func TestGetDockerTagsWithDate(t *testing.T) {

	dockerSvc := DockerService{}
	dockerSvc.httpClient = &HTTPClientMock{}

	dockerTagRequest := model.ListDockerTagsRequest{}
	dockerTagRequest.From = "2006-01-02"
	dockerTagRequest.ImageName = "repoOne/my_image"

	dockerRepo := model.DockerRepo{}
	dockerRepo.Password = "123"
	dockerRepo.Username = "abc"
	dockerRepo.Host = "http://myrepository.com.br"

	dockerDaoMock := mocks.DockerDAOInterface{}
	dockerDaoMock.On("GetDockerRepositoryByHost", mock.Anything).Return(dockerRepo, nil)
	globalCache := sync.Map{}

	result, e := dockerSvc.GetDockerTagsWithDate(dockerTagRequest, &dockerDaoMock, &globalCache)
	assert.Nil(t, e)
	assert.NotNil(t, result)
	assert.Equal(t, 2, len(result.TagResponse))

}

func TestBuilder(t *testing.T) {
	ds := DockerServiceBuilder()
	assert.NotNil(t, ds)
}

func TestGetHTTPClient(t *testing.T) {
	c := getHTTPClient()
	assert.NotNil(t, c)
}

func TestDoRequest(t *testing.T) {
	dockerSvc := DockerService{}
	dockerSvc.httpClient = &HTTPClientImpl{}
	bytes, e := dockerSvc.httpClient.doRequest("http://google.com.br", "alfa", "beta")
	assert.Nil(t, e)
	assert.NotNil(t, bytes)
}

func TestDefineTagResponseFromCache(t *testing.T) {

	cacheInfo := CacheInfo{}
	cacheInfo.globalCache = &sync.Map{}
	cacheInfo.dateFrom = time.Now()
	cacheInfo.matchFromDate = false
	tagResult := model.ListDockerTagsResult{}
	tagResult.TagResponse = make([]model.TagResponse, 0)

	cacheInfo.result = &tagResult

	repo := model.DockerRepo{}
	cacheInfo.repo = &repo
	cacheInfo.imageName = "image"

	defineTagResponseFromCache("docker.io/alfa", "latest", time.Now(), cacheInfo)

	assert.NotNil(t, cacheInfo.result.TagResponse)

}

func TestDefineTagResponseFromCacheMatch(t *testing.T) {

	date2 := time.Now()
	date1 := time.Now()

	cacheInfo := CacheInfo{}
	cacheInfo.globalCache = &sync.Map{}
	cacheInfo.globalCache.Store("docker.io/alfa", date1)

	cacheInfo.dateFrom = time.Now()
	cacheInfo.matchFromDate = true
	tagResult := model.ListDockerTagsResult{}
	tagResult.TagResponse = make([]model.TagResponse, 0)

	cacheInfo.result = &tagResult

	repo := model.DockerRepo{}
	cacheInfo.repo = &repo
	cacheInfo.imageName = "image"

	defineTagResponseFromCache("docker.io/alfa", "latest", date2, cacheInfo)

	assert.NotNil(t, cacheInfo.result.TagResponse)

}

func TestDefineTagResponseFromCacheMatchNot(t *testing.T) {

	cacheInfo := CacheInfo{}
	cacheInfo.globalCache = &sync.Map{}

	cacheInfo.dateFrom = time.Now()
	cacheInfo.matchFromDate = true
	tagResult := model.ListDockerTagsResult{}
	tagResult.TagResponse = make([]model.TagResponse, 0)

	cacheInfo.result = &tagResult

	repo := model.DockerRepo{}
	cacheInfo.repo = &repo
	cacheInfo.imageName = "image"

	defineTagResponseFromCache("docker.io/alfa", "latest", time.Now(), cacheInfo)

	assert.NotNil(t, cacheInfo.result.TagResponse)

}
