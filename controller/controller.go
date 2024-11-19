package controller

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type Handler func(Event)

type Controller struct {
	informers []*Informer
	handler   func(Event)
	ticker    *time.Ticker
	buffer    map[string]Event
	cache     chan Event
}

func New(handler Handler, selectors []Selector, resync, flush time.Duration) (*Controller, error) {
	s := &Controller{
		informers: make([]*Informer, 0),
		handler:   handler,
		ticker:    time.NewTicker(flush),
		buffer:    make(map[string]Event),
		cache:     make(chan Event),
	}
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	for _, selector := range selectors {
		informer, err := NewInformer(selector, client, resync, s.cache)
		if err != nil {
			log.Warn().Err(err).Msgf("unable to create informer for '%s'", selector.GroupVersionKind())
			continue
		}
		s.informers = append(s.informers, informer)
	}
	return s, nil
}

func (s *Controller) flush() {
	for key, evt := range s.buffer {
		s.handler(evt)
		delete(s.buffer, key)
	}
	log.Debug().Msg("Controller Flushed...")
}

func (s *Controller) process(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	hasFlushed := false
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("Controller Stopping...")
			s.flush()
			wg.Done()
			return
		case evt := <-s.cache:
			if hasFlushed {
				log.Debug().Msgf("Controller buffered %s ...", evt.Key)
				s.buffer[evt.Key] = evt
			} else {
				s.handler(evt)
			}
		case <-s.ticker.C:
			hasFlushed = true
			s.flush()
		}
	}
}

func (s *Controller) Run(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Msgf("Controller Starting...")
	go s.process(ctx, wg)
	for _, informer := range s.informers {
		go informer.Run(ctx, wg)
	}
}
