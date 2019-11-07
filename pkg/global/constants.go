//+build !test

package global

const (
	//Function id of a method log
	Function string = "function"

	//KubeConfigBasePath  - Path of kubeconfig directory
	KubeConfigBasePath string = "./config/"

	//HelmDir - Path of Helm diretory
	HelmDir = "./.helm/"

	//AccessDenied AccessDenied
	AccessDenied = "Acccess Denied"

	//ParameterIDError ParameterIDError
	ParameterIDError = "Error processing parameter id: "

	//ContentType ContentType
	ContentType = "Content-Type"

	//JSONContentType JSONContentType
	JSONContentType = "application/json; charset=UTF-8"
)
