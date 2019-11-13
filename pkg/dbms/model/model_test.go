package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDepAnalyseMarshal(t *testing.T) {

	expected := "{\"nodes\":[{\"id\":\"1\",\"color\":\"blue\",\"symbolType\":" +
		"\"cube\",\"svg\":\"\"}],\"links\":[{\"source\":\"serviceA\",\"target\":\"serviceB\"}]}"

	links := make([]DepLink, 1)
	nodes := make([]Node, 1)

	links[0].Source = "serviceA"
	links[0].Target = "serviceB"
	nodes[0].ID = "1"
	nodes[0].Color = "blue"
	nodes[0].SymbolType = "cube"
	nodes[0].Svg = ""

	result := &DepAnalyse{
		Nodes: nodes,
		Links: links,
	}

	out, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	stringResult := string(out)

	if stringResult != expected {
		t.Errorf("Error getting node name: got %v want %v", string(out), expected)
	}

}

func TestModel(t *testing.T) {
	mp := ConfigMap{}
	mp.Name = "alfa"
	mp.Value = "xpto"
	assert.NotNil(t, mp)
}
