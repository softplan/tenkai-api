package util

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"encoding/json"
)

func GetHttpBody(r *http.Request) ([]byte, error) {
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

//UnmarshalPayload
func UnmarshalPayload(r *http.Request, payload interface{}) error {
	body, error := GetHttpBody(r)
	if error != nil {
		return error
	}
	if error = json.Unmarshal(body, &payload); error != nil {
		return error
	}
	return nil
}