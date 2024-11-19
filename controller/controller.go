package controller

import (
	"context"
	"github.com/rs/zerolog/log"
	"slices"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

type Handler func(Event)

type Controller struct {
	client   *Client
	informer cache.SharedIndexInformer
	handler  func(Event)
	ticker   *time.Ticker
	buffer   map[string]Event
	cache    chan Event
}

func New(handler Handler, selector Selector, resync, flush time.Duration) (*Controller, error) {
	s := &Controller{
		handler: handler,
		ticker:  time.NewTicker(flush),
		buffer:  make(map[string]Event),
		cache:   make(chan Event),
	}
	client, err := NewClient(selector)
	if err != nil {
		return nil, err
	}
	s.client = client
	s.informer = client.NewInformerFactory(resync)
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onCreate,
		UpdateFunc: s.onUpdate,
		DeleteFunc: s.onDelete,
	}
	if _, err := s.informer.AddEventHandler(handlers); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Controller) key(item *unstructured.Unstructured) string {
	gvk := s.client.GVK
	return strings.Join(
		slices.DeleteFunc(
			[]string{gvk.Group, gvk.Version, gvk.Kind, item.GetNamespace(), string(item.GetUID())},
			func(s string) bool { return s == "" },
		),
		"/")
}

func (s *Controller) onCreate(obj any) {
	s.cache <- Event{
		Op:  OpCreate,
		Old: obj.(*unstructured.Unstructured),
		New: obj.(*unstructured.Unstructured),
	}
}

func (s *Controller) onUpdate(oldObj, newObj any) {
	s.cache <- Event{
		Op:  OpUpdate,
		Old: oldObj.(*unstructured.Unstructured),
		New: newObj.(*unstructured.Unstructured),
	}
}

func (s *Controller) onDelete(obj any) {
	s.cache <- Event{
		Op:  OpDelete,
		Old: obj.(*unstructured.Unstructured),
		New: obj.(*unstructured.Unstructured),
	}
}

func (s *Controller) flush() {
	for key, evt := range s.buffer {
		s.handler(evt)
		delete(s.buffer, key)
	}
	log.Debug().Msgf("[%s] Controller Flushed...", s.client.ID())
}

func (s *Controller) process(ctx context.Context) {
	hasFlushed := false
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("[%s] Controller Stopping...", s.client.ID())
			s.flush()
			return
		case evt := <-s.cache:
			if hasFlushed {
				log.Debug().Msgf("[%s] Controller buffered %s ...", s.client.ID(), s.key(evt.New))
				s.buffer[s.key(evt.New)] = evt
			} else {
				s.handler(evt)
			}
		case <-s.ticker.C:
			hasFlushed = true
			s.flush()
		}
	}
}

func (s *Controller) Run(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Msgf("[%s] Controller Starting...", s.client.ID())
	wg.Add(1)
	go s.process(ctx)
	s.informer.Run(ctx.Done())
	time.Sleep(1 * time.Second)
	wg.Done()
	log.Info().Msgf("[%s] Controller Stopped.", s.client.ID())
}
