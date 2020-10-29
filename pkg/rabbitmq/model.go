package rabbitmq

import (
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
)

//PayloadRabbit consumer
type PayloadRabbit struct {
	UpgradeRequest helmapi.UpgradeRequest `json:"upgradeRequest"`
	Name           string `json:"name"`
	Token          string `json:"token"`
	Filename       string `json:"filename"`
	CACertificate  string `json:"ca_certificate"`
	ClusterURI     string `json:"cluster_uri"`
	Namespace      string `json:"namespace"`
}