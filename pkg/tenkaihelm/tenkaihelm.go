package tenkaihelm

import (
	"net/http"

	"github.com/softplan/tenkai-api/pkg/util"
)

//HelmAPIInteface interface
type HelmAPIInteface interface {
	DoGetRequest(url string) (responseBytes []byte, err error)
	//doPostRequest(url string) (resp *http.Response, err error)
}

//HelmAPIImpl struct
type HelmAPIImpl struct {
}

//DoGetRequest make a get request
func (helm HelmAPIImpl) DoGetRequest(url string) (responseBytes []byte, err error) {
	resp, err := http.Get(url)
	if err != nil{
		return responseBytes, err
	}
	responseBytes, err = util.GetHTTPBodyResponse(resp)
	if err != nil{
		return []byte{}, err
	}
	return responseBytes,nil
}
