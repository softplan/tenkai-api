package analyser

import (
	"testing"
)

func TestGetNodeName(t *testing.T) {
	expected := "alfa:1.0"
	result := getNodeName("alfa", "1.0")
	if result != expected {
		t.Errorf("Error getting node name: got %v want %v", result, expected)
	}
}
