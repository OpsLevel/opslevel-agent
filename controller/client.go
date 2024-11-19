package controller

import (
	"fmt"
	"slices"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	client dynamic.Interface
	GVR    schema.GroupVersionResource
	GVK    schema.GroupVersionKind
}

// NewClient
// This creates a wrapper which gives you an initialized and connected kubernetes client
// It then has a number of helper functions
func NewClient(selector Selector) (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	gv, err := schema.ParseGroupVersion(selector.ApiVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid apiVersion: %w", err)
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    selector.Kind,
	}

	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gv.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GVK to GVR: %w", err)
	}

	return &Client{
		client: client,
		GVR:    mapping.Resource,
		GVK:    gvk,
	}, nil
}

func (s *Client) NewInformerFactory(resync time.Duration) cache.SharedIndexInformer {
	return dynamicinformer.NewFilteredDynamicSharedInformerFactory(s.client, resync, corev1.NamespaceAll, nil).ForResource(s.GVR).Informer()
}

func (s *Client) ID() string {
	return strings.Join(
		slices.DeleteFunc(
			[]string{s.GVK.Group, s.GVK.Version, s.GVK.Kind},
			func(s string) bool { return s == "" },
		),
		"/")
}
