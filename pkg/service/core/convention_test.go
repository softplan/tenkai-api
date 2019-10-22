package core

import (
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetKubeConfigFileName(t *testing.T) {
	conv := &ConventionImpl{}
	name := conv.GetKubeConfigFileName("alfa", "beta")
	assert.Equal(t, name, global.KubeConfigBasePath+"alfa_beta")
}
