package ddd

import (
	"fmt"
	"github.com/startcodextech/goevents/registry"
)

type (
	EventSetter interface {
		setEvents([]Event)
	}
)

func SetEvents(events ...Event) registry.BuildOption {
	return func(v interface{}) error {
		if agg, ok := v.(EventSetter); ok {
			agg.setEvents(events)
			return nil
		}
		return fmt.Errorf("%T does not have the method setEvents([]ddd.Event)", v)
	}
}
