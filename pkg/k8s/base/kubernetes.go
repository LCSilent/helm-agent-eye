package base

import (
	openapi_v2 "github.com/google/gnostic/openapiv2"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Kubernetes struct {
	Config                      *rest.Config
	Clientset                   kubernetes.Interface
	DiscoveryClient             discovery.DiscoveryInterface
	DynamicClient               dynamic.Interface
	DeferredDiscoveryRESTMapper *restmapper.DeferredDiscoveryRESTMapper
	OpenapiSchema               *openapi_v2.Document
	MetricsClient               metricsclientset.Interface
	ApiextensionsClient         apiextensionsclient.Interface
}

// RestConfig returns the underlying rest.Config.
func (k *Kubernetes) RestConfig() *rest.Config {
	return k.Config
}

// NewKubernetes creates a new Kubernetes client
func NewKubernetes() (*Kubernetes, error) {
	config, clientset, err := newK8SClient()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	apiextClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		Config:                      config,
		Clientset:                   clientset,
		DiscoveryClient:             discoveryClient,
		DynamicClient:               dynamicClient,
		DeferredDiscoveryRESTMapper: restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)),
		OpenapiSchema:               &openapi_v2.Document{},
		MetricsClient:               metricsClient,
		ApiextensionsClient:         apiextClient,
	}, nil
}
