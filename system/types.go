package system

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/startcodextech/goevents/config"
	"github.com/startcodextech/goevents/waiter"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type (
	Service interface {
		Config() config.AppConfig
		SqlDB() *sql.DB
		MongoDB() *mongo.Client
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
