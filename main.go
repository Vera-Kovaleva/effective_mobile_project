package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ef_project/internal/infra/log"
	"ef_project/internal/infra/noerr"

	"ef_project/internal/domain"

	"ef_project/internal/adapters/database"
	"ef_project/internal/infra/repository"

	httpapi "ef_project/internal/adapters/http"
	oapi "ef_project/internal/generated/oapi"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
	"golang.org/x/sync/errgroup"
)

const (
	exitOK = iota
	exitDotEnvFailed
	exitServersFailed
)

const (
	readimeout        = 100 * time.Millisecond
	readHeaderTimeout = 100 * time.Millisecond
)

func main() {
	os.Exit(Run(context.Background()))
}

func Run(ctx context.Context) int {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		slog.ErrorContext(ctx, "Loading environment variables failed.", log.ErrorAttr(err))

		return exitDotEnvFailed
	}

	var stop context.CancelFunc
	ctx, stop = signal.NotifyContext(
		ctx,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	switch gin.Mode() {
	case gin.DebugMode:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	provider := database.NewPostgresProvider(
		noerr.Must(pgxpool.New(ctx, os.Getenv("DB_CONNECTION"))),
	)
	defer provider.Close()

	router := gin.Default()

	subscriptionsService := domain.NewSubscriptionService(
		provider,
		repository.NewSubscription(),
	)

	middlewares := []oapi.StrictMiddlewareFunc{
		func(f strictgin.StrictGinHandlerFunc, _ string) strictgin.StrictGinHandlerFunc {
			return func(ctx *gin.Context, request any) (any, error) {
				log.SetRequestID(ctx, uuid.NewString())

				return f(ctx, request)
			}
		},
	}

	oapi.RegisterHandlers(
		router,
		oapi.NewStrictHandler(
			httpapi.NewServer(subscriptionsService),
			middlewares,
		),
	)

	var eg errgroup.Group
	startHTTPServer(ctx, &eg, router)

	if err := eg.Wait(); err != nil {
		slog.ErrorContext(ctx, "Runing servers failed.", log.ErrorAttr(err))

		return exitServersFailed
	}

	return exitOK
}

func startHTTPServer(ctx context.Context, eg *errgroup.Group, router *gin.Engine) {
	httpSrv := &http.Server{
		Addr:              os.Getenv("HTTP_ADDRESS"),
		Handler:           router,
		ReadTimeout:       readimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	eg.Go(func() error {
		slog.InfoContext(ctx, "Starting HTTP server", slog.String("addr", httpSrv.Addr))
		err := httpSrv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}

		return err
	})
	eg.Go(func() error {
		<-ctx.Done()

		return httpSrv.Shutdown(ctx)
	})
}
