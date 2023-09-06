package eventsource

import (
	"fmt"
	"github.com/start-codex/goevents/ddd"
)

type (
	EventApplier interface {
		ApplyEvent(event ddd.Event) error
	}

	EventCommitter interface {
		CommitEvents()
	}
)

func LoadEvent(v interface{}, event ddd.AggregateEvent) error {
	type loader interface {
		EventApplier
		VersionSetter
	}

	agg, ok := v.(loader)
	if !ok {
		return fmt.Errorf("%T does not have the methods implemented to load events", v)
	}

	if err := agg.ApplyEvent(event); err != nil {
		return err
	}
	agg.SetVersion(event.AggregateVersion())

	return nil
}
