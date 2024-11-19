package controller

import (
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Operation string

const (
	OpCreate Operation = "create"
	OpUpdate Operation = "update"
	OpDelete Operation = "delete"
)

type Event struct {
	Key string
	Op  Operation
	Old *unstructured.Unstructured
	New *unstructured.Unstructured
}

func (s Event) ExternalKind() string {
	gvr := s.New.GroupVersionKind()
	return strings.Join(
		slices.DeleteFunc(
			[]string{gvr.Group, gvr.Version, gvr.Kind},
			func(s string) bool { return s == "" },
		),
		"/")
}

func (s Event) ExternalID(cluster string) string {
	return strings.Join(
		slices.DeleteFunc(
			[]string{cluster, s.New.GetNamespace(), string(s.New.GetUID())},
			func(s string) bool { return s == "" },
		),
		"/")
}
