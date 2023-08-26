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

func (bus *QueryEvents) Publish(ctx context.Context, query goevents.Query) (goevents.Payload, error) {
	handler, exists := bus.handlers[query.Name]
	if !exists {
		return nil, goevents.ErrNoExitsTopic
	}

	errChan := make(chan error, 1)
	result := make(chan []byte, 1)

	go func(ctx context.Context, query goevents.Query, queryHandler goevents.QueryHandler) {
		payload, err := queryHandler(ctx, query)
		if err != nil {
			result <- nil
			errChan <- err
			return
		}
		subscribers, existsSubscribers := bus.subscribers[query.Name]
		if existsSubscribers {
			for _, subscriber := range subscribers {
				go func(ctx context.Context, query goevents.Query, payload goevents.Payload, handler goevents.QuerySubscribeHandler) {
					handler(ctx, query, payload)
				}(ctx, query, payload, subscriber)
			}
		}
		result <- payload
		errChan <- nil
	}(ctx, query, handler)

	err := <-errChan
	payload := <-result
	close(errChan)
	close(result)

	return payload, err
}

func (bus *QueryEvents) RegisterHandler(query string, handler goevents.QueryHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.handlers == nil {
		bus.handlers = make(map[string]goevents.QueryHandler)
	}

	bus.handlers[query] = handler
}
