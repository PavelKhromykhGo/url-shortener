package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/analytics"
	"github.com/PavelKhromykhGo/url-shortener/internal/config"
	"github.com/PavelKhromykhGo/url-shortener/internal/kafka"
	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/internal/storage/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	kafkago "github.com/segmentio/kafka-go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logg := logger.New(cfg.Env)
	logg.Info("starting analytics consumer",
		logger.String("env", cfg.Env),
		logger.String("kafka_topic", cfg.KafkaClicksTopic),
	)

	pgpool, err := pgxpool.New(ctx, cfg.PostgresDSN)
	if err != nil {
		logg.Fatal("failed to connect to postgres", logger.Error(err))
	}
	defer pgpool.Close()

	if err := pgpool.Ping(ctx); err != nil {
		logg.Fatal("failed to ping postgres", logger.Error(err))
	}
	analyticsRepo := postgres.NewAnalyticsRepository(pgpool)
	analyticsService := analytics.NewService(analyticsRepo, logg)

	groupID := getEnv("KAFKA_CLICKS_CONSUMER_GROUP", "clicks-analytics-consumer")

	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  cfg.KafkaBrokers,
		Topic:    cfg.KafkaClicksTopic,
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			logg.Warn("failed to close kafka reader", logger.Error(err))
		}
	}()

	logg.Info("analytics-consumer started",
		logger.String("group_id", groupID),
	)

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		for {
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					logg.Info("context canceled, stopping consumer loop")
					return
				}

				logg.Error("failed to read kafka message", logger.Error(err))
				errCh <- err
				return
			}

			logg.Debug("kafka message received",
				logger.Int64("offset", msg.Offset),
				logger.String("partition", string(rune(msg.Partition))),
			)

			var event kafka.ClickEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				logg.Error("failed to unmarshal kafka message", logger.Error(err),
					logger.Error(err),
				)
				continue
			}

			eventCtx, cancelEvent := context.WithTimeout(ctx, 5*time.Second)

			if err := analyticsService.ProcessClick(eventCtx, event); err != nil {
				logg.Error("failed to process click event", logger.Error(err),
					logger.Int64("link_id", event.LinkID),
					logger.String("short_code", event.ShortCode),
				)
			}

			cancelEvent()
		}
	}()

	select {
	case sig := <-sigCh:
		logg.Info("shutdown signal received", logger.String("signal", sig.String()))
		cancel()
	case err := <-errCh:
		if err != nil {
			logg.Error("analytics consumer failed", logger.Error(err))
		}
		cancel()
	}

	shutdownCtx, shutsownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutsownCancel()

	<-shutdownCtx.Done()
	logg.Info("analytics consumer stopped")

}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
