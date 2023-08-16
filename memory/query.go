package memory

import (
	"context"
	"github.com/start-codex/goevents"
	"sync"
)

type (
	QueryEvents struct {
		mu          sync.Mutex
		handlers    map[string]goevents.QueryHandler
		subscribers map[string][]goevents.QuerySubscribeHandler
	}
)

func (bus *QueryEvents) Subscribe(query string, handler goevents.QuerySubscribeHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.subscribers == nil {
		bus.subscribers = make(map[string][]goevents.QuerySubscribeHandler)
	}

	subscribers, exists := bus.subscribers[query]
	if !exists {
		subscribers = make([]goevents.QuerySubscribeHandler, 0)
	}

	subscribers = append(subscribers, handler)

	bus.subscribers[query] = subscribers
}

func (bus *QueryEvents) Publish(ctx context.Context, query goevents.Query) error {
	wg := sync.WaitGroup{}
	handler, exists := bus.handlers[query.Name]
	if !exists {
		return goevents.ErrNoExitsTopic
	}

	errChan := make(chan error, 1)

	wg.Add(1)
	go func(ctx context.Context, query goevents.Query, queryHandler goevents.QueryHandler) {
		defer wg.Done()
		payload, err := queryHandler(ctx, query)
		if err != nil {
			errChan <- err
			return
		}
		subscribers, existsSubscribers := bus.subscribers[query.Name]
		if existsSubscribers {
			for _, subscriber := range subscribers {
				wg.Add(1)
				go func(ctx context.Context, query goevents.Query, payload goevents.Payload, handler goevents.QuerySubscribeHandler) {
					defer wg.Done()
					handler(ctx, query, payload)
				}(ctx, query, payload, subscriber)
			}
		}
	}(ctx, query, handler)

	wg.Wait()
	close(errChan)

	return <-errChan
}

func (bus *QueryEvents) RegisterHandler(query string, handler goevents.QueryHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.handlers == nil {
		bus.handlers = make(map[string]goevents.QueryHandler)
	}

	bus.handlers[query] = handler
}
