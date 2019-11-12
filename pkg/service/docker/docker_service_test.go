package dockerapi

import (
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
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