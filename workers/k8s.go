package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/opslevel/opslevel-go/v2024"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"opslevel-agent/controller"
)

type K8SWorker struct {
	client      *opslevel.Client
	selectors   []controller.Selector
	cluster     string
	integration string
}

func NewK8SWorker(cluster string, integration string, selectors []controller.Selector, client *opslevel.Client) *K8SWorker {
	return &K8SWorker{
		client:      client,
		selectors:   selectors,
		cluster:     cluster,
		integration: integration,
	}
}

func (s *K8SWorker) Run(ctx context.Context, wg *sync.WaitGroup, resync, flush time.Duration) {
	for _, selector := range s.selectors {
		ctrl, err := controller.New(s.handle, selector, resync, flush)
		if err != nil {
			log.Error().Err(err).
				Str("api", selector.ApiVersion).
				Str("kind", selector.Kind).
				Msgf("failed to start k8s controller")
			continue
		}
		go ctrl.Run(ctx, wg)
		time.Sleep(50 * time.Millisecond) // Jitter
	}
}

func (s *K8SWorker) handle(evt controller.Event) {
	kind := evt.ExternalKind()
	id := evt.ExternalID(s.cluster)

	switch evt.Op {
	case controller.OpCreate, controller.OpUpdate:
		value, err := s.parse(evt.New)
		if err != nil {
			log.Error().Err(err).Msgf("failed to convert k8s resource")
		}
		s.sendUpsert(kind, id, value)
	case controller.OpDelete:
		s.sendDelete(kind, id)
	}
}

func (s *K8SWorker) parse(item *unstructured.Unstructured) (opslevel.JSON, error) {
	unstructured.RemoveNestedField(item.Object, "metadata", "managedFields")
	// unstructured.RemoveNestedField(item.Object, "spec")
	// unstructured.RemoveNestedField(item.Object, "status", "conditions")

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
		log.Info().Msgf("[DRYRUN] UPSERT %s | %s", kind, id)
		log.Debug().Msgf("\t%#v", value)
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
