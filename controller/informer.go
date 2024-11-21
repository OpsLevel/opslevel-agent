package controller

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"slices"
	"strings"
	"sync"
	"time"
)

type Informer struct {
	selector Selector
	informer cache.SharedIndexInformer
	queue    chan<- Event
}

func NewInformer(selector Selector, client *Client, resync time.Duration, queue chan<- Event) (*Informer, error) {
	mapping, err := client.mapper.RESTMapping(selector.GroupKind(), selector.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve GVK to GVR: %w", err)
	}
	informer := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client.client, resync, corev1.NamespaceAll, nil).ForResource(mapping.Resource).Informer()
	s := &Informer{
		selector: selector,
		informer: informer,
		queue:    queue,
	}
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onCreate,
		UpdateFunc: s.onUpdate,
		DeleteFunc: s.onDelete,
	}
	_, err = informer.AddEventHandler(handlers)
	return s, err
}

func (s *Informer) key(item *unstructured.Unstructured) string {
	return strings.Join(
		slices.DeleteFunc(
			[]string{s.selector.Group, s.selector.Version, s.selector.Kind, item.GetNamespace(), string(item.GetUID())},
			func(s string) bool { return s == "" },
		),
		"/")
}

func (s *Informer) onCreate(obj any) {
	s.queue <- Event{
		Key: s.key(obj.(*unstructured.Unstructured)),
		Op:  OpCreate,
		Old: obj.(*unstructured.Unstructured),
		New: obj.(*unstructured.Unstructured),
	}
}

func (s *Informer) onUpdate(oldObj, newObj any) {
	s.queue <- Event{
		Key: s.key(newObj.(*unstructured.Unstructured)),
		Op:  OpUpdate,
		Old: oldObj.(*unstructured.Unstructured),
		New: newObj.(*unstructured.Unstructured),
	}
}

func (s *Informer) onDelete(obj any) {
	s.queue <- Event{
		Key: s.key(obj.(*unstructured.Unstructured)),
		Op:  OpDelete,
		Old: obj.(*unstructured.Unstructured),
		New: obj.(*unstructured.Unstructured),
	}
}

func (s *Informer) Run(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Msgf("[%s] Informer Starting...", s.selector.GroupVersionKind())
	wg.Add(1)
	defer wg.Done()
	s.informer.Run(ctx.Done())
	log.Info().Msgf("[%s] Informer Stopping...", s.selector.GroupVersionKind())
}
