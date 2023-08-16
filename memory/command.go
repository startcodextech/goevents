package memory

import (
	"context"
	"github.com/start-codex/goevents"
	"sync"
)

type (
	CommandEvents struct {
		mu          sync.Mutex
		handlers    map[string]goevents.CommandHandler
		subscribers map[string][]goevents.CommandSubscribeHandler
	}
)

func (bus *CommandEvents) Subscribe(command string, handler goevents.CommandSubscribeHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.subscribers == nil {
		bus.subscribers = make(map[string][]goevents.CommandSubscribeHandler)
	}

	subscribers, exists := bus.subscribers[command]
	if !exists {
		subscribers = make([]goevents.CommandSubscribeHandler, 0)
	}

	subscribers = append(subscribers, handler)

	bus.subscribers[command] = subscribers
}

func (bus *CommandEvents) Publish(ctx context.Context, command goevents.Command) error {
	wg := sync.WaitGroup{}
	handler, exists := bus.handlers[command.Name]
	if !exists {
		return goevents.ErrNoExitsTopic
	}

	errChan := make(chan error, 1)

	wg.Add(1)
	go func(ctx context.Context, command goevents.Command, commandHandler goevents.CommandHandler) {
		defer wg.Done()
		err := commandHandler(ctx, command)
		if err != nil {
			errChan <- err
		}
		subscribers, existsSubscribers := bus.subscribers[command.Name]
		if existsSubscribers {
			for _, subscriber := range subscribers {
				wg.Add(1)
				go func(ctx context.Context, command goevents.Command, handler goevents.CommandSubscribeHandler) {
					defer wg.Done()
					handler(ctx, command)
				}(ctx, command, subscriber)
			}
		}
	}(ctx, command, handler)

	wg.Wait()
	close(errChan)

	return <-errChan
}

func (bus *CommandEvents) RegisterHandler(query string, handler goevents.CommandHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if bus.handlers == nil {
		bus.handlers = make(map[string]goevents.CommandHandler)
	}

	bus.handlers[query] = handler
}
