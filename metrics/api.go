package metrics

import "github.com/prometheus/client_golang/prometheus"

// APIMetrics holds Prometheus metrics for the API.
type APIMetrics struct {
	InFlight prometheus.GaugeVec

	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec

	LinksCreatedTotal *prometheus.CounterVec
	RedirectsTotal    *prometheus.CounterVec

	KafkaProduceTotal    *prometheus.CounterVec
	KafkaProduceDuration *prometheus.HistogramVec
}

// NewAPIMetrics creates and registers API metrics with the given Prometheus registerer.
func NewAPIMetrics(req prometheus.Registerer) *APIMetrics {
	m := &APIMetrics{
		InFlight: *prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "api_in_flight_requests",
				Help: "Number of in-flight API requests",
			},
			[]string{"route"},
		),
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"route", "method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_request_duration_seconds",
				Help:    "Duration of API requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"route", "method"},
		),
		LinksCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_links_created_total",
				Help: "Total number of links created via the API",
			},
			[]string{"result"},
		),

		RedirectsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_redirects_total",
				Help: "Total number of redirects via the API",
			},
			[]string{"result"},
		),

		KafkaProduceTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_kafka_produce_total",
				Help: "Total number of Kafka produce attempts",
			},
			[]string{"topic", "result"},
		),

		KafkaProduceDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "api_kafka_produce_duration_seconds",
				Help:    "Duration of Kafka produce operations",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"topic"},
		),
	}
	req.MustRegister(
		m.InFlight,
		m.RequestsTotal,
		m.RequestDuration,
		m.LinksCreatedTotal,
		m.RedirectsTotal,
		m.KafkaProduceTotal,
		m.KafkaProduceDuration,
	)
	return m
}
