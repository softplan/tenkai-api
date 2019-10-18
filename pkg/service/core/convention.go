package core

import "github.com/softplan/tenkai-api/pkg/global"

//ConventionInterface ConventionInterface
type ConventionInterface interface {
	GetKubeConfigFileName(group string, name string) string
}

//ConventionImpl ConventionImpl
type ConventionImpl struct {
}

//GetKubeConfigFileName GetKubeConfigFileName
func (c ConventionImpl) GetKubeConfigFileName(group string, name string) string {
	return global.KubeConfigBasePath + group + "_" + name
}
