package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

type Operation string

const (
	OpCreate Operation = "create"
	OpUpdate Operation = "update"
	OpDelete Operation = "delete"
)

func nullHandler(item *unstructured.Unstructured)                  {}
func nullChangedHandler(oldOjb, newObj *unstructured.Unstructured) {}

type Controller struct {
	OnAdd     func(obj *unstructured.Unstructured)
	OnUpdate  func(obj *unstructured.Unstructured)
	OnChanged func(oldObj, newObj *unstructured.Unstructured)
	OnDelete  func(obj *unstructured.Unstructured)

	informer cache.SharedIndexInformer
}

func New(selector Selector, resync time.Duration) (*Controller, error) {
	client, err := NewK8SClient()
	if err != nil {
		return nil, err
	}
	gvr, err := client.GetGVR(selector)
	if err != nil {
		return nil, err
	}
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client.Dynamic, resync, corev1.NamespaceAll, nil)
	return (&Controller{
		OnAdd:     nullHandler,
		OnUpdate:  nullHandler,
		OnChanged: nullChangedHandler,
		OnDelete:  nullHandler,

		informer: factory.ForResource(*gvr).Informer(),
	}).setup()
}

func (s *Controller) setup() (*Controller, error) {
	_, err := s.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			s.OnAdd(obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, newObj any) {
			s.OnUpdate(newObj.(*unstructured.Unstructured))
			s.OnChanged(oldObj.(*unstructured.Unstructured), newObj.(*unstructured.Unstructured))
		},
		DeleteFunc: func(obj any) {
			s.OnDelete(obj.(*unstructured.Unstructured))
		},
	})

	return s, err
}

func (s *Controller) Run(ctx context.Context) {
	s.informer.Run(ctx.Done())
}
