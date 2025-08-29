package config

import (
	"opslevel-agent/controller"
)

var (
	DefaultConfiguration = &Configuration{
		Selectors: []controller.Selector{
			{
				ApiVersion: "v1",
				Kind:       "Namespace",
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "StatefulSet",
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
			},
			{
				ApiVersion: "networking.k8s.io/v1",
				Kind:       "Ingress",
			},
			{
				ApiVersion: "v1",
				Kind:       "Service",
			},
			{
				ApiVersion: "batch/v1",
				Kind:       "CronJob",
			},
		},
	}
)

type Configuration struct {
	Selectors []controller.Selector `json:"selectors" yaml:"selectors"`
}
