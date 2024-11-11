package workers

import (
	"context"
	"encoding/json"
	"github.com/opslevel/opslevel-go/v2024"
	k8s "github.com/opslevel/opslevel-k8s-controller/v2024"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type K8SOperation string

const (
	K8SCreate K8SOperation = "create"
	K8SUpdate K8SOperation = "update"
	K8SDelete K8SOperation = "delete"
)

type K8SEvent struct {
	Op           K8SOperation
	ExternalType string
	ExternalID   string
	Data         opslevel.JSON
}

type K8SWorker struct {
	// TODO: Cluster Name
	selectors []k8s.K8SSelector
}

func NewK8SWorker() *K8SWorker {
	return &K8SWorker{
		selectors: []k8s.K8SSelector{
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Namespaces: []string{"default"},
			},
		},
	}
}

func (s *K8SWorker) Run(ctx context.Context, wg *sync.WaitGroup) {
	queue := make(chan K8SEvent)
	for _, selector := range s.selectors {
		s.producer(ctx, wg, selector, queue)
	}
	go s.consumer(ctx, wg, queue)
}

func (s *K8SWorker) producer(ctx context.Context, wg *sync.WaitGroup, selector k8s.K8SSelector, queue chan<- K8SEvent) {
	controller, err := k8s.NewK8SController(selector, 24*time.Hour)
	if err != nil {
		log.Error().Err(err).
			Str("api", selector.ApiVersion).
			Str("kind", selector.Kind).
			Msgf("failed to start k8s controller")
		return
	}
	controller.OnAdd = parser(K8SCreate, queue)
	controller.OnUpdate = parser(K8SUpdate, queue)
	controller.OnDelete = parser(K8SDelete, queue)
	wg.Add(1)
	controller.Start(ctx, wg)
}

func (s *K8SWorker) consumer(ctx context.Context, wg *sync.WaitGroup, queue <-chan K8SEvent) {
	wg.Add(1)
	log.Info().Msgf("starting consumer")
	//client := opslevel.NewGQLClient()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("stopping consumer")
			wg.Done()
			return
		case r := <-queue:
			log.Info().Msgf("HERE !!! %v", r.Data.ToJSON())
		}
	}
}

func parser(op K8SOperation, queue chan<- K8SEvent) func(any) {
	return func(item any) {
		// TODO: convert to a k8s Metadata
		// TODO: parse out the API Group and Kind into ExternalType
		// TODO: parse out the Name, Namespace and Cluster into ExternalID
		data, err := json.Marshal(item)
		if err != nil {
			log.Error().Err(err).Msgf("failed to marshal k8s resource")
			return
		}
		j, err := opslevel.NewJSON(string(data))
		if err != nil {
			log.Error().Err(err).Msgf("failed to create opslevel JSON")
			return
		}
		queue <- K8SEvent{
			Op:           op,
			ExternalType: "apps/v1/Deployment",
			ExternalID:   "dev/opslevel/web",
			Data:         *j,
		}
	}
}
