package controller

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type K8SClient struct {
	Client  kubernetes.Interface
	Dynamic dynamic.Interface
	Mapper  *restmapper.DeferredDiscoveryRESTMapper
}

// NewK8SClient
// This creates a wrapper which gives you an initialized and connected kubernetes client
// It then has a number of helper functions
func NewK8SClient() (*K8SClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	client1, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client2, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// Suppress k8s client-go logs
	klog.SetLogger(logr.Discard())
	return &K8SClient{Client: client1, Dynamic: client2, Mapper: mapper}, nil
}

func (c *K8SClient) GetMapping(selector Selector) (*meta.RESTMapping, error) {
	gv, gvErr := schema.ParseGroupVersion(selector.ApiVersion)
	if gvErr != nil {
		return nil, gvErr
	}
	gvk := gv.WithKind(selector.Kind)

	mapping, mappingErr := c.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if mappingErr != nil {
		return nil, mappingErr
	}

	return mapping, nil
}

func (c *K8SClient) GetGVR(selector Selector) (*schema.GroupVersionResource, error) {
	mapping, err := c.GetMapping(selector)
	if err != nil {
		return nil, err
	}
	return &mapping.Resource, nil
}
