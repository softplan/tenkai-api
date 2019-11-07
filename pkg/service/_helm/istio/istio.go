package istio

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

//ClientInterface ClientInterface
type ClientInterface interface {
	RESTClient() rest.Interface
}

//Client Client
type Client struct {
	restClient rest.Interface
}

//NewForConfig NewForConfig
func NewForConfig(c *rest.Config) (*Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

//NewForConfigOrDie NewForConfigOrDie
func NewForConfigOrDie(c *rest.Config) *Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

//New New
func New(c rest.Interface) *Client {
	return &Client{c}
}

//IstioSchemeGroupVersion IstioSchemeGroupVersion
var IstioSchemeGroupVersion = schema.GroupVersion{Group: "networking.istio.io", Version: "v1alpha3"}

func setConfigDefaults(config *rest.Config) error {
	gv := IstioSchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

//RESTClient RESTClient
func (c *Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
