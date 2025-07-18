package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/opslevel/opslevel-go/v2025"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"opslevel-agent/controller"
)

type K8SWorker struct {
	cluster     string
	integration string
	client      *opslevel.Client
	rest        *resty.Client
}

func NewK8SWorker(ctx context.Context, wg *sync.WaitGroup, cluster string, integration string, selectors []controller.Selector, client *opslevel.Client, rest *resty.Client, resync, flush time.Duration) {
	controller.Run(ctx, wg, selectors, resync, flush, &K8SWorker{
		cluster:     cluster,
		integration: integration,
		client:      client,
		rest:        rest,
	})
}

func (s *K8SWorker) Handle(evt controller.Event) {
	kind := evt.ExternalKind()
	id := evt.ExternalID(s.cluster)

	if strings.Contains(s.integration, "integrations/custom/webhook") {
		switch evt.Op {
		case controller.OpCreate, controller.OpUpdate:
			s.sendEvent(kind, id, evt.New)
		}
	} else {
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

func (s *K8SWorker) sendEvent(kind string, id string, value *unstructured.Unstructured) {
	kind = strings.Replace(kind, "/", "_", -1)
	if viper.GetBool("dry-run") {
		log.Info().Msgf("[DRYRUN] POST %s | %s", kind, id)
		log.Debug().Msgf("\t%#v", value)
	} else {
		url := fmt.Sprintf("%s?external_kind=%s", s.integration, kind)
		resp, err := s.rest.R().SetBody(value).Post(url)
		if err != nil {
			log.Error().Err(err).Msgf("error during post")
			return
		}
		if resp.StatusCode() > 299 {
			log.Error().Msgf("%v", resp)
			return
		}
		log.Info().Msgf("POST %s | %s", kind, id)
	}
}

func (s *K8SWorker) sendUpsert(kind string, id string, value opslevel.JSON) {
	var m struct {
		Payload struct {
			Errors []opslevel.Error
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
			Errors []opslevel.Error
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
