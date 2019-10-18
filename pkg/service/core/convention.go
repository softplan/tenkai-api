package core

import "github.com/softplan/tenkai-api/pkg/global"

type ConventionInterface interface {
	GetKubeConfigFileName(group string, name string) string
}

type ConventionImpl struct {
}

func (c ConventionImpl) GetKubeConfigFileName(group string, name string) string {
	return global.KubeConfigBasePath + group + "_" + name
}
