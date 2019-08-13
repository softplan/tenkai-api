package util

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
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
