package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/Sushka21/url-shortener/internal/config"
	"github.com/Sushka21/url-shortener/internal/controller"
	"github.com/Sushka21/url-shortener/internal/repository/inmemory"
	"github.com/Sushka21/url-shortener/internal/repository/postgres"
	urlUsecase "github.com/Sushka21/url-shortener/internal/usecase/urlshortener"
	db "github.com/Sushka21/url-shortener/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"go.uber.org/zap"
)

func Run(logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var repo urlUsecase.Repository
	var dbpool *pgxpool.Pool
	defer func() {
		if dbpool != nil {
			logger.Info("closing postgres connection pool...")
			dbpool.Close()
		}
	}()

	switch cfg.StorageType {
	case "postgres":
		logger.Info("initializing postgres storage...")
		pgxcfg, err := pgxpool.ParseConfig(cfg.ConstructPostgresURL())
		if err != nil {
			logger.Error("can not create pgxpool cfg", zap.Error(err))
			return err
		}
		pgxcfg.MaxConns = 8
		pgxcfg.MinConns = 1
		pgxcfg.HealthCheckPeriod = config.HealthCheckPeriod
		pgxcfg.MaxConnLifetime = 0
		pgxcfg.MaxConnIdleTime = config.MaxConnIdleTime

		var errPool error
		dbpool, errPool = pgxpool.NewWithConfig(ctx, pgxcfg)
		if errPool != nil {
			logger.Error("can not create pgxpool", zap.Error(errPool))
			return errPool
		}

		if errSetup := db.SetupPostgres(dbpool, logger); errSetup != nil {
			logger.Error("can not creat migrtions", zap.Error(errSetup))
			return errSetup
		}

		repo = postgres.NewPostgresRepository(dbpool)

	case "inmemory", "":
		logger.Info("initializing in-memory storage...")
		repo = inmemory.NewInMemoryRepository()

	default:
		return fmt.Errorf("unknown storage type specified: %s", cfg.StorageType)
	}
	urlService := urlUsecase.NewURLService(repo, logger)
	ctrl := controller.New(urlService, logger, cfg.ConstructBaseURL())

	mux := http.NewServeMux()
	ctrl.Register(mux)

	wgerr, ctx := errgroup.WithContext(ctx)
	wgerr.Go(func() error {
		return runHTTPServer(ctx, logger, cfg, mux)
	})

	if err := wgerr.Wait(); err != nil {
		logger.Error("failed to wait", zap.Error(err))
		return err
	}

	return nil
}

func runHTTPServer(ctx context.Context, logger *zap.Logger, cfg *config.Config, mux *http.ServeMux) error {
	srv := &http.Server{
		Addr:    net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: mux,
	}
	shutdownDone := make(chan struct{})
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
		defer cancel()
		logger.Info("shutting down http gateway")

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("http shutdown error", zap.Error(err))
		}
		close(shutdownDone)
	}()

	err := srv.ListenAndServe()

	<-shutdownDone

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
