package system

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/start-codex/goevents/config"
	"github.com/start-codex/goevents/waiter"
	"google.golang.org/grpc"
)

type (
	Service interface {
		Config() config.AppConfig
		DB() *sql.DB
		JS() nats.JetStreamContext
		Mux() *chi.Mux
		RPC() *grpc.Server
		Waiter() waiter.Waiter
		Logger() zerolog.Logger
	}
	Module interface {
		Startup(context.Context, Service) error
	}
)
