package configs

import (
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestReadConfigNotFound(t *testing.T) {
	config, error := ReadConfig("notfoundfile")
	if error == nil || config != nil {
		t.Error("Error - Config file does not exists but ReadConfig was ok")
	}
}
