package config

import (
	"opslevel-agent/controller"
)

var (
	DefaultConfiguration = &Configuration{
		Selectors: []controller.Selector{
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
			},
		},
	}
	ExtendedConfiguration = &Configuration{
		Selectors: []controller.Selector{
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
				ApiVersion: "batch/v1",
				Kind:       "CronJob",
			},
			{
				ApiVersion: "v1",
				Kind:       "Service",
			},
			{
				ApiVersion: "apiregistration.k8s.io/v1",
				Kind:       "APIService",
			},
			{
				ApiVersion: "autoscaling/v2",
				Kind:       "HorizontalPodAutoscaler",
			},
			{
				ApiVersion: "v1",
				Kind:       "Pod",
			},
		},
	}
)

type Configuration struct {
	Selectors []controller.Selector `json:"selectors" yaml:"selectors"`
}
