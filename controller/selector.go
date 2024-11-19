package controller

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"slices"
	"strings"
)

type Selector struct {
	ApiVersion string `json:"apiVersion" yaml:"apiVersion"`
	Group      string `json:"-" yaml:"-"`
	Version    string `json:"-" yaml:"-"`
	Kind       string `json:"kind" yaml:"kind"`

	ready bool
}

func (s *Selector) parse() {
	if s.ready {
		return
	}
	gv, err := schema.ParseGroupVersion(s.ApiVersion)
	if err != nil {
		panic(fmt.Errorf("invalid apiVersion: %w", err))
	}
	s.Group = gv.Group
	s.Version = gv.Version
	s.ready = true
}

func (s Selector) ID() string {
	s.parse()
	return strings.Join(
		slices.DeleteFunc(
			[]string{s.Group, s.Version, s.Kind},
			func(s string) bool { return s == "" },
		),
		"/")
}

func (s Selector) GroupKind() schema.GroupKind {
	s.parse()
	return schema.GroupKind{Group: s.Group, Kind: s.Kind}
}

func (s Selector) GroupVersionKind() schema.GroupVersionKind {
	s.parse()
	return schema.GroupVersionKind{Group: s.Group, Version: s.Version, Kind: s.Kind}
}
