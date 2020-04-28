package httpsvc

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
)

//HTTPServiceBuilder HTTPServiceBuilder
func HTTPServiceBuilder() *HTTPService {
	r := &HTTPService{}
	r.httpClient = HTTPClientImpl{}
	return r
}

//HTTPClient HTTPClient
type HTTPClient interface {
	post(url string, payload *bytes.Buffer) ([]byte, error)
}

//HTTPClientImpl HTTPClientImpl
type HTTPClientImpl struct {
	HTTPClient
}

func (h HTTPClientImpl) post(url string, payload *bytes.Buffer) ([]byte, error) {

	client := getHTTPClient()

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	return bodyBytes, err
}

// HTTPService is used to concretize HTTPServiceInterface
type HTTPService struct {
	httpClient HTTPClient
}

// HTTPServiceInterface can be used to make http requests
type HTTPServiceInterface interface {
	Post(url string, payload *bytes.Buffer) ([]byte, error)
}

func getHTTPClient() *http.Client {
	tr := &http.Transport{}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}

//Post Post
func (svc HTTPService) Post(url string, payload *bytes.Buffer) ([]byte, error) {
	return svc.httpClient.post(url, payload)
}
