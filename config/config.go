package config

import k8s "github.com/opslevel/opslevel-k8s-controller/v2024"

var (
	DefaultConfiguration = &Configuration{
		Selectors: []k8s.K8SSelector{
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
		},
	}
	ExtendedConfiguration = &Configuration{
		Selectors: []k8s.K8SSelector{
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "StatefulSet",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "DaemonSet",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "batch/v1",
				Kind:       "CronJob",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "v1",
				Kind:       "Service",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "v1",
				Kind:       "APIService",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "autoscaling/v2",
				Kind:       "HorizontalPodAutoscaler",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
			{
				ApiVersion: "v1",
				Kind:       "Pod",
				Excludes:   []string{".metadata.namespace == \"kube-system\""},
			},
		},
	}
)

type Configuration struct {
	Selectors []k8s.K8SSelector
}
