package workers

import (
	"context"
	"encoding/json"
	"github.com/spf13/viper"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/opslevel/opslevel-go/v2024"
	k8s "github.com/opslevel/opslevel-k8s-controller/v2024"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"opslevel-agent/controller"
)

type K8SWorker struct {
	client      *opslevel.Client
	selectors   []k8s.K8SSelector
	cluster     string
	integration string
}

func NewK8SWorker(cluster string, integration string, selectors []k8s.K8SSelector, client *opslevel.Client) *K8SWorker {
	return &K8SWorker{
		client:      client,
		selectors:   selectors,
		cluster:     cluster,
		integration: integration,
	}
}

func (s *K8SWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	for _, selector := range s.selectors {
		go s.producer(ctx, wg, selector)
	}
}

func (s *K8SWorker) producer(ctx context.Context, wg *sync.WaitGroup, selector k8s.K8SSelector) {
	controller, err := controller.New(selector, 24*time.Hour)
	if err != nil {
		log.Error().Err(err).
			Str("api", selector.ApiVersion).
			Str("kind", selector.Kind).
			Msgf("failed to start k8s controller")
		return
	}
	controller.OnAdd = s.parser(true)
	controller.OnUpdate = s.parser(true)
	controller.OnDelete = s.parser(false)
	wg.Add(1)
	controller.Run(ctx)
	wg.Done()
}

func (s *K8SWorker) parser(update bool) func(*unstructured.Unstructured) {
	return func(item *unstructured.Unstructured) {
		gvr := item.GroupVersionKind()
		// I'm sorry to whomever future person reads this code because of a bug and hates me for this
		kind := strings.Join(
			slices.DeleteFunc(
				[]string{gvr.Group, gvr.Version, gvr.Kind},
				func(s string) bool { return s == "" },
			),
			"/")
		id := strings.Join(
			slices.DeleteFunc(
				[]string{s.cluster, item.GetNamespace(), string(item.GetUID())},
				func(s string) bool { return s == "" },
			),
			"/")

		if update {
			value, err := s.funcName(item)
			if err != nil {
				log.Error().Err(err).Msgf("failed to convert k8s resource")
				return
			}
			s.sendUpsert(kind, id, value)
		} else {
			s.sendDelete(kind, id)
		}
	}
}

func (s *K8SWorker) funcName(item *unstructured.Unstructured) (opslevel.JSON, error) {
	// TODO: Cleanup Data Based on known types - Deployment, Statefulset, Daemonset
	unstructured.RemoveNestedField(item.Object, "metadata", "managedFields")
	//unstructured.RemoveNestedField(item.Object, "spec")
	//unstructured.RemoveNestedField(item.Object, "status", "conditions")
	return s.toJson(item)
}

func (s *K8SWorker) toJson(item any) (opslevel.JSON, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	var data opslevel.JSON
	if err = json.Unmarshal(b, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *K8SWorker) sendUpsert(kind string, id string, value opslevel.JSON) {
	var m struct {
		Payload struct {
			Errors []opslevel.OpsLevelErrors
		} `graphql:"integrationSourceObjectUpsert(externalId: $id, externalKind: $kind, integration: $integration, value: $value)"`
	}
	v := opslevel.PayloadVariables{
		"kind":        kind,
		"id":          id,
		"integration": *opslevel.NewIdentifier(s.integration),
		"value":       value,
	}
	if viper.GetBool("dry-run") {
		log.Info().Msgf("[DRYRUN] UPSERT %s | %s | %#v", kind, id, value)
	} else {
		log.Info().Msgf("UPSERT %s | %s", kind, id)
		err := s.client.Mutate(&m, v, opslevel.WithName("IntegrationSourceObjectUpsert"))
		if err != nil {
			log.Error().Err(err).Msgf("error during upsert mutate")
		}
	}
}

func (s *K8SWorker) sendDelete(kind string, id string) {
	var m struct {
		Payload struct {
			Errors []opslevel.OpsLevelErrors
		} `graphql:"integrationSourceObjectDelete(externalId: $id, externalKind: $kind, integration: $integration)"`
	}
	v := opslevel.PayloadVariables{
		"kind":        kind,
		"id":          id,
		"integration": *opslevel.NewIdentifier(s.integration),
	}
	if viper.GetBool("dry-run") {
		log.Info().Msgf("[DRYRUN] DELETE %s | %s ", kind, id)
	} else {
		log.Info().Msgf("DELETE %s | %s ", kind, id)
		err := s.client.Mutate(&m, v, opslevel.WithName("IntegrationSourceObjectDelete"))
		if err != nil {
			log.Error().Err(err).Msgf("error during delete mutate")
		}
	}
}
