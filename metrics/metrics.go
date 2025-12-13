package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once sync.Once

	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec

	KafkaProducerPublishedTotal *prometheus.CounterVec
	KafkaProducerErrorsTotal    *prometheus.CounterVec

	KafkaConsumerProcessedTotal  *prometheus.CounterVec
	KafkaConsumerProcessDuration *prometheus.HistogramVec
	KafkaConsumerLagSeconds      *prometheus.HistogramVec
)

func MustInit(serviceName string) {
	once.Do(func() {
		HTTPRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "route", "status"},
		)

		HTTPRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_request_duration_seconds",
				Help: "Duration of HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		)

		KafkaProducerPublishedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_producer_published_total",
				Help: "Total number of messages published by Kafka producer",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"topic"},
		)

		KafkaProducerErrorsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_producer_errors_total",
				Help: "Total number of errors encountered by Kafka producer",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"topic"},
		)

		KafkaConsumerProcessedTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_consumer_processed_total",
				Help: "Total number of messages processed by Kafka consumer",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"topic", "result"},
		)

		KafkaConsumerProcessDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "kafka_consumer_process_duration_seconds",
				Help: "Duration of Kafka consumer message processing",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
				Buckets: prometheus.DefBuckets,
			},
			[]string{"topic"},
		)

		KafkaConsumerLagSeconds = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "kafka_consumer_lag_seconds",
				Help: "Lag of Kafka consumer in seconds",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
			},
			[]string{"topic"},
		)

		prometheus.MustRegister(
			HTTPRequestsTotal,
			HTTPRequestDuration,
			KafkaProducerPublishedTotal,
			KafkaProducerErrorsTotal,
			KafkaConsumerProcessedTotal,
			KafkaConsumerProcessDuration,
			KafkaConsumerLagSeconds,
		)

	})
}
