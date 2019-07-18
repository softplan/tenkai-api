package util

import "testing"

func TestGetMultipleValues(t *testing.T) {
	keywords := GetReplacebleKeyName("http://${SERVER}:${PORT}/teste/abc/${LATEST}")
	if !(keywords[0] == "SERVER" && keywords[1] == "PORT" && keywords[2] == "LATEST") {
		t.Fatal("Fail")
	}
}

func TestGetSingleValues(t *testing.T) {
	keywords := GetReplacebleKeyName("http://${SERVER}:8080")
	if !(keywords[0] == "SERVER") {
		t.Fatal("Fail")
	}
}
