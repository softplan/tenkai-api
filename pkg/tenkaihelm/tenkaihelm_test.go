package tenkaihelm

import (
	"testing"
)


func TestDoRequestOk(test *testing.T) {
	helm := HelmAPIImpl{}
	helm.DoGetRequest("https://google.com.br")
}

func TestDoRequestFail(test *testing.T) {
	helm := HelmAPIImpl{}
	helm.DoGetRequest("xpto")
}