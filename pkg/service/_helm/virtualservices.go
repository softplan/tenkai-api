package helmapi

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/service/istio"
	"time"
)

func (svc HelmServiceImpl) GetVirtualServices(kubeconfig string, namespace string) ([]string, error) {

	hostNames := make([]string, 0)

	restConfig, _, err := svc.GetHelmConnection().GetKubeClient("", kubeconfig)
	if err != nil {
		return nil, err
	}

	ix, err := istio.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	c := ix.RESTClient()

	var timeout time.Duration

	bytes, _ := c.Get().
		Namespace(namespace).
		Resource("virtualservices").
		Timeout(timeout).
		Do().Raw()

	var objmap map[string]interface{}
	json.Unmarshal(bytes, &objmap)

	items := objmap["items"].([]interface{})
	for _, e := range items {
		switch e.(type) {
		case map[string]interface{}:
			k := e.(map[string]interface{})
			if k["spec"] != "" {
				hostName := ""
				specMap := k["spec"].(map[string]interface{})
				if specMap["hosts"] != "" {
					hosts := specMap["hosts"].([]interface{})
					for _, h := range hosts {
						hostName = h.(string)
					}
				}
				if specMap["http"] != "" {
					http := specMap["http"].([]interface{})
					for _, k := range http {
						httpMap := k.(map[string]interface{})
						uri := httpMap["match"]
						matchArray := uri.([]interface{})
						for _, matchElement := range matchArray {
							prefixName := ""
							uriValue := matchElement.(map[string]interface{})["uri"]
							if uriValue != nil {
								prefix := uriValue.(map[string]interface{})["prefix"]
								if prefix != nil {
									prefixName = prefix.(string)
								}
							}
							hostName = hostName + prefixName
						}
					}
				}
				hostNames = append(hostNames, hostName)
			}
		}
	}

	return hostNames, nil

}
