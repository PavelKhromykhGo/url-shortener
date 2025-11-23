package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/config"
	"github.com/PavelKhromykhGo/url-shortener/internal/httpapi"
	"github.com/PavelKhromykhGo/url-shortener/internal/httpserver"
	"github.com/PavelKhromykhGo/url-shortener/internal/id"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/shortener"
	"github.com/PavelKhromykhGo/url-shortener/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logg := logger.New(cfg.Env)
	logg.Info("starting api server",
		logger.String("env", cfg.Env),
		logger.String("addr", cfg.HTTPAddr))

	pgPool, err := pgxpool.New(ctx, cfg.PostgresDNS)
	if err != nil {
		logg.Fatal("failed to connect to postgres", logger.Error(err))
	}

	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		logg.Fatal("failed to ping postgres", logger.Error(err))
	}

	linksRepo := postgres.NewLinksRepository(pgPool)

	idGen := id.NewRandomGenerator(8)

	shortenerService := shortener.NewService(shortener.Config{
		BaseURL:   cfg.BaseURL,
		LinksRepo: linksRepo,
		LinkCache: nil,
		IDGen:     idGen,
		Logger:    logg,
	})

	deps := httpapi.Deps{
		Logger:           logg,
		ShortenerService: shortenerService,
	}

	router := httpapi.NewRouter(deps)

	srv := httpserver.New(httpserver.Config{
		Addr:         cfg.HTTPAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Logger:       logg,
	})

	go func() {
		if err := srv.Start(); err != nil {
			logg.Fatal("failed to start http server", logger.Error(err))
		}
	}()

	logg.Info("api server started")

	sig := <-sigCh
	logg.Info("shutdown signal received", logger.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		logg.Error("failed to stop http server gracefully", logger.Error(err))
	} else {
		logg.Info("http server stopped gracefully")
	}

	logg.Info("api server stopped")
}
