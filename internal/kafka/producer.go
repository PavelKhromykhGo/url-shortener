package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/PavelKhromykhGo/url-shortener/internal/logger"
	"github.com/PavelKhromykhGo/url-shortener/metrics"
	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

// ClickEvent represents a click event for a shortened URL.
type ClickEvent struct {
	SchemaVersion int32     `json:"schema_version"`
	EventType     string    `json:"event_type"`
	EventID       string    `json:"event_id"`
	LinkID        int64     `json:"link_id"`
	ShortCode     string    `json:"short_code"`
	ClickedAt     time.Time `json:"clicked_at"`
	UserAgent     string    `json:"user_agent,omitempty"`
	Referer       string    `json:"referer,omitempty"`
	IP            string    `json:"ip,omitempty"`
}

// NewClickEvent creates a new ClickEvent with the provided details.
func NewClickEvent(linkID int64, shortCode, userAgent, referer, ip string, clickedAt time.Time) ClickEvent {
	return ClickEvent{
		SchemaVersion: 1,
		EventType:     "click",
		EventID:       uuid.New().String(),
		LinkID:        linkID,
		ShortCode:     shortCode,
		ClickedAt:     clickedAt.UTC(),
		UserAgent:     userAgent,
		Referer:       referer,
		IP:            ip,
	}

}

// ClickProducer defines the interface for publishing click events to Kafka.
type ClickProducer interface {
	PublishClick(ctx context.Context, event ClickEvent) error
	Close() error
}

// clickProducer is an implementation of ClickProducer using Kafka.
type clickProducer struct {
	writer *kafkago.Writer
	logger logger.Logger
}

// PublishClick publishes a click event to Kafka.
func (p clickProducer) PublishClick(ctx context.Context, event ClickEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal click event: %w", err)
	}

	key := []byte(strconv.FormatInt(event.LinkID, 10))

	msg := kafkago.Message{
		Key:   key,
		Value: data,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		metrics.KafkaProducerErrorsTotal.WithLabelValues(p.writer.Topic).Inc()
		p.logger.Error("failed to publish click event to kafka",
			logger.Error(err),
			logger.Int64("link_id", event.LinkID),
			logger.String("short_code", event.ShortCode),
		)
		return err
	}

	metrics.KafkaProducerPublishedTotal.WithLabelValues(p.writer.Topic).Inc()

	p.logger.Debug("click event published to kafka",
		logger.Int64("link_id", event.LinkID),
		logger.String("short_code", event.ShortCode),
	)
	return nil
}

// Close closes the Kafka writer.
func (p clickProducer) Close() error {
	return p.writer.Close()
}

// NewClickProducer creates a new ClickProducer with the given Kafka brokers and topic.
func NewClickProducer(brokers []string, topic string, log logger.Logger) (ClickProducer, error) {
	if len(brokers) == 0 {
		return nil, fmt.Errorf("no kafka brokers provided")
	}
	if topic == "" {
		return nil, fmt.Errorf("no kafka topic provided")
	}

	w := &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.Hash{},
		RequiredAcks: kafkago.RequireAll,
		Async:        false,
	}

	return &clickProducer{
		writer: w,
		logger: log,
	}, nil
}
