package controller

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type Handler interface {
	Handle(Event)
}

type Controller struct {
	handler Handler
	ticker  *time.Ticker
	buffer  map[string]Event
	cache   chan Event
}

func (s *Controller) flush() {
	for key, evt := range s.buffer {
		s.handler.Handle(evt)
		delete(s.buffer, key)
	}
	log.Debug().Msg("Controller Flushed...")
}

func (s *Controller) process(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	hasFlushed := false
	for {
		select {
		case <-ctx.Done():
			s.flush()
			return
		case evt := <-s.cache:
			if hasFlushed {
				log.Debug().Msgf("Controller buffered %s ...", evt.Key)
				s.buffer[evt.Key] = evt
			} else {
				s.handler.Handle(evt)
			}
		case <-s.ticker.C:
			hasFlushed = true
			s.flush()
		}
	}
}

func Run(ctx context.Context, wg *sync.WaitGroup, selectors []Selector, resync, flush time.Duration, handler Handler) {
	s := &Controller{
		handler: handler,
		ticker:  time.NewTicker(flush),
		buffer:  make(map[string]Event),
		cache:   make(chan Event),
	}
	client, err := NewClient()
	if err != nil {
		log.Warn().Err(err).Msg("unable to create kubernetes client")
		return
	}

	for _, selector := range selectors {
		informer, err := NewInformer(selector, client, resync, s.cache)
		if err != nil {
			log.Warn().Err(err).Msgf("unable to create informer for '%s'", selector.GroupVersionKind())
			continue
		}
		go informer.Run(ctx, wg)
	}
	log.Info().Msgf("Controller Starting...")
	s.process(ctx, wg)
	log.Info().Msgf("Controller Stopping...")
}
