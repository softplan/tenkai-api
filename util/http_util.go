package util

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
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