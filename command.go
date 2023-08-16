package goevents

import (
	"context"
	"github.com/start-codex/utils/id"
	"time"
)

type (
	Params interface{}

	Command struct {
		MessageID     string    `json:"message_id"`
		CorrelationID string    `json:"correlation_id"`
		Name          string    `json:"Name"`
		Created       time.Time `json:"time"`
		Params        Params
	}

	CommandHandler          func(ctx context.Context, command Command) error
	CommandSubscribeHandler func(ctx context.Context, command Command)

	CommandBus interface {
		Subscribe(command string, handler CommandSubscribeHandler)
		Publish(ctx context.Context, command Command) error
		RegisterHandler(command string, handler CommandHandler)
	}
)

func CreateCommand(name string, payload Params) Command {
	mID, _ := id.New()
	cID, _ := id.New()
	return Command{
		MessageID:     "mid:" + mID.String(),
		CorrelationID: "cid:" + cID.String(),
		Name:          name,
		Created:       time.Now(),
		Params:        payload,
	}
}

func CreateCommandCorrelationID(uuid, name string, params Params) Command {
	uid, _ := id.New()
	return Command{
		MessageID:     "mid:" + uid.String(),
		CorrelationID: "cid:" + uuid,
		Name:          name,
		Created:       time.Now(),
		Params:        params,
	}
}
