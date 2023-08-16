package memory

import (
	"github.com/start-codex/goevents"
	"sync"
)

type (
	MemoryEvents struct {
		Command goevents.CommandBus
		Query   goevents.QueryBus
	}
)

func CreateMemoryEvents() *MemoryEvents {
	return &MemoryEvents{
		Command: &CommandEvents{
			mu:          sync.Mutex{},
			handlers:    make(map[string]goevents.CommandHandler),
			subscribers: make(map[string][]goevents.CommandSubscribeHandler),
		},
		Query: &QueryEvents{
			mu:          sync.Mutex{},
			handlers:    make(map[string]goevents.QueryHandler),
			subscribers: make(map[string][]goevents.QuerySubscribeHandler),
		},
	}
}
