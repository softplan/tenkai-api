package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/dbms/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

const loadBalancerWidth = 16

//GetServices - Get Service information
func GetServices(kubeconfig string, namespace string) ([]model.Service, error) {
	services := make([]model.Service, 0)
	_, client, err := getKubeClient(settings.KubeContext, kubeconfig)
	if err != nil {
		return services, err
	}
	opts := &metav1.ListOptions{}
	list, err := client.CoreV1().Services(namespace).List(*opts)

	var service *model.Service
	for _, element := range list.Items {
		service = fillService(element)
		services = append(services, *service)
	}
	return services, nil
}

func makePortString(ports []v1.ServicePort) string {
	pieces := make([]string, len(ports))
	for ix := range ports {
		port := &ports[ix]
		pieces[ix] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		if port.NodePort > 0 {
			pieces[ix] = fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
		}
	}
	return strings.Join(pieces, ",")
}

func fillService(service v1.Service) *model.Service {
	result := model.Service{Name: service.Name}
	result.ClusterIP = service.Spec.ClusterIP
	result.ExternalIP = getServiceExternalIP(&service, false)
	result.Ports = makePortString(service.Spec.Ports)
	result.Type = string(service.Spec.Type)
	result.Age = translateTimestampSince(service.CreationTimestamp)
	return &result
}

func getServiceExternalIP(svc *v1.Service, wide bool) string {
	switch svc.Spec.Type {
	case v1.ServiceTypeClusterIP:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case v1.ServiceTypeNodePort:
		if len(svc.Spec.ExternalIPs) > 0 {
			return strings.Join(svc.Spec.ExternalIPs, ",")
		}
		return "<none>"
	case v1.ServiceTypeLoadBalancer:
		lbIps := loadBalancerStatusStringer(svc.Status.LoadBalancer, wide)
		if len(svc.Spec.ExternalIPs) > 0 {
			results := []string{}
			if len(lbIps) > 0 {
				results = append(results, strings.Split(lbIps, ",")...)
			}
			results = append(results, svc.Spec.ExternalIPs...)
			return strings.Join(results, ",")
		}
		if len(lbIps) > 0 {
			return lbIps
		}
		return "<pending>"
	case v1.ServiceTypeExternalName:
		return svc.Spec.ExternalName
	}
	return "<unknown>"
}

func loadBalancerStatusStringer(s v1.LoadBalancerStatus, wide bool) string {
	ingress := s.Ingress
	result := sets.NewString()
	for i := range ingress {
		if ingress[i].IP != "" {
			result.Insert(ingress[i].IP)
		} else if ingress[i].Hostname != "" {
			result.Insert(ingress[i].Hostname)
		}
	}

	r := strings.Join(result.List(), ",")
	if !wide && len(r) > loadBalancerWidth {
		r = r[0:(loadBalancerWidth-3)] + "..."
	}
	return r
}
