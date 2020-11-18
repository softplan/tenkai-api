package util

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//GetHTTPBody - Returns body
func GetHTTPBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatalln("Error on body", err)
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error - body closed", err)
	}
	return body, nil
}

//GetHTTPBodyResponse - Returns body
func GetHTTPBodyResponse(r *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		log.Fatalln("Error on body", err)
		return nil, err
	}
	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error - body closed", err)
	}
	return body, nil
}

//UnmarshalPayload - Transform a raw post request into a struct
func UnmarshalPayload(r *http.Request, payload interface{}) error {
	body, error := GetHTTPBody(r)
	if error != nil {
		return error
	}
	if error = json.Unmarshal(body, &payload); error != nil {
		return error
	}
	return nil
}

//GetPrincipal - Returns principal from request
func GetPrincipal(r *http.Request) model.Principal {
	var principal model.Principal
	principalString := r.Header.Get("principal")
	json.Unmarshal([]byte(principalString), &principal)
	return principal
}
