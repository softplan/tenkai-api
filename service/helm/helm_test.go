package helmapi

import (
	"testing"
)

func TestSetupConnectionWithTiller(t *testing.T) {
	settings.TillerHost = "alfa"
	err := setupConnection()
	if err != nil {
		t.Errorf("Handler returned wrong status code: got %v want %v", err, nil)
	}
}
