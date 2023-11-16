package system

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
	"github.com/pressly/goose/v3"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/config"
	"github.com/startcodextech/goevents/logger"
	"github.com/startcodextech/goevents/waiter"
	"go.mongodb.org/mongo-driver/mongo"
	options2 "go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io/fs"
	"net"
	"net/http"
	"time"
)

type (
	System struct {
		cfg     config.AppConfig
		db      *sql.DB
		mongodb *mongo.Client
		nc      *nats.Conn
		js      nats.JetStreamContext
		mux     *chi.Mux
		rpc     *grpc.Server
		waiter  waiter.Waiter
		logger  zerolog.Logger
		tp      *sdktrace.TracerProvider
	}
)

func NewSystem(cfg config.AppConfig) (*System, error) {
	s := &System{cfg: cfg}

	s.initWaiter()

	if err := s.initDB(); err != nil {
		return nil, err
	}

	if err := s.initJS(); err != nil {
		return nil, err
	}

	if err := s.initOpenTelemetry(); err != nil {
		return nil, err
	}

	s.initMux()
	s.initRpc()
	s.initLogger()

	return s, nil
}

func (s *System) initOpenTelemetry() error {
	exporter, err := otlptracegrpc.New(context.Background())
	if err != nil {
		return err
	}

	s.tp = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(s.tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	s.waiter.Cleanup(func() {
		if err := s.tp.Shutdown(context.Background()); err != nil {
			s.logger.Error().Err(err).Msg("ran into an issue shutting down the tracer provider")
		}
	})

	return nil
}

func (s *System) Config() config.AppConfig {
	return s.cfg
}

func (s *System) initDB() (err error) {
	switch s.cfg.DB.Driver {
	case "mongo":
		options := options2.Client().ApplyURI(s.cfg.DB.Conn)
		s.mongodb, err = mongo.Connect(context.Background(), options)
		return err
	default:
		s.db, err = sql.Open(s.cfg.DB.Driver, s.cfg.DB.Conn)
		return err
	}
}

func (s *System) MigrateDB(fs fs.FS) error {
	if s.cfg.DB.Driver == "mongo" {
		return fmt.Errorf("It is not allowed because a SQL database is not managed")
	}
	goose.SetBaseFS(fs)
	if err := goose.SetDialect("sql"); err != nil {
		return err
	}
	if err := goose.Up(s.db, "."); err != nil {
		return err
	}
	return nil
}

// SqlDB returns a pointer to the sql.DB instance associated with the System.
// This method provides access to the SQL database client, which can be used
// for executing SQL queries and operations. It is important to ensure that
// the System instance has been properly initialized and connected to the
// database before calling this method.
func (s *System) SqlDB() *sql.DB {
	return s.db
}

// MongoDB returns a pointer to the mongo.Client instance associated with the System.
// This method provides access to the MongoDB client, allowing for operations
// such as querying and manipulating documents within MongoDB collections. Similar
// to DB(), it is crucial to ensure that the System has been appropriately initialized
// and connected to a MongoDB instance before invoking this method.
func (s *System) MongoDB() *mongo.Client {
	return s.mongodb
}

func (s *System) initJS() (err error) {
	s.nc, err = nats.Connect(s.cfg.Nats.URL)
	if err != nil {
		return err
	}
	s.js, err = s.nc.JetStream()
	if err != nil {
		return err
	}

	_, err = s.js.AddStream(&nats.StreamConfig{
		Name:     s.cfg.Nats.Stream,
		Subjects: []string{fmt.Sprintf("%s.>", s.cfg.Nats.Stream)},
	})

	return err
}

func (s *System) JS() nats.JetStreamContext {
	return s.js
}

func (s *System) initLogger() {
	s.logger = logger.New(logger.LogConfig{
		Environment: s.cfg.Environment,
		LogLevel:    logger.Level(s.cfg.LogLevel),
	})
}

func (s *System) Logger() zerolog.Logger {
	return s.logger
}

func (s *System) initMux() {
	s.mux = chi.NewMux()
	s.mux.Use(middleware.Heartbeat("/liveness"))
	s.mux.Method("GET", "/metrics", promhttp.Handler())
}

func (s *System) Mux() *chi.Mux {
	return s.mux
}

func (s *System) initRpc() {
	s.rpc = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			otelgrpc.UnaryServerInterceptor(),
			serverErrorUnaryInterceptor(),
		),
		// If there are streaming endpoints also add
		// grpc.StreamInterceptor(
		// 	otelgrpc.StreamServerInterceptor(),
		// ),
	)
	reflection.Register(s.rpc)
}

func (s *System) RPC() *grpc.Server {
	return s.rpc
}

func (s *System) initWaiter() {
	s.waiter = waiter.New(waiter.CatchSignals())
}

func (s *System) Waiter() waiter.Waiter {
	return s.waiter
}

func (s *System) WaitForWeb(ctx context.Context) error {
	webServer := &http.Server{
		Addr:    s.cfg.Web.Address(),
		Handler: s.mux,
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Printf("web server started; listening at http://localhost%s\n", s.cfg.Web.Port)
		defer fmt.Println("web server shutdown")
		if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-gCtx.Done()
		fmt.Println("web server to be shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		if err := webServer.Shutdown(ctx); err != nil {
			return err
		}
		return nil
	})

	return group.Wait()
}

func (s *System) WaitForRPC(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.cfg.Rpc.Address())
	if err != nil {
		return err
	}

	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Println("rpc server started")
		defer fmt.Println("rpc server shutdown")
		if err := s.RPC().Serve(listener); err != nil && err != grpc.ErrServerStopped {
			return err
		}
		return nil
	})
	group.Go(func() error {
		<-gCtx.Done()
		fmt.Println("rpc server to be shutdown")
		stopped := make(chan struct{})
		go func() {
			s.RPC().GracefulStop()
			close(stopped)
		}()
		timeout := time.NewTimer(s.cfg.ShutdownTimeout)
		select {
		case <-timeout.C:
			// Force it to stop
			s.RPC().Stop()
			return fmt.Errorf("rpc server failed to stop gracefully")
		case <-stopped:
			return nil
		}
	})

	return group.Wait()
}

func (s *System) WaitForStream(ctx context.Context) error {
	closed := make(chan struct{})
	s.nc.SetClosedHandler(func(*nats.Conn) {
		close(closed)
	})
	group, gCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		fmt.Println("message stream started")
		defer fmt.Println("message stream stopped")
		<-closed
		return nil
	})
	group.Go(func() error {
		<-gCtx.Done()
		return s.nc.Drain()
	})
	return group.Wait()
}

func serverErrorUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)
		return resp, errors.SendGRPCError(err)
	}
}
