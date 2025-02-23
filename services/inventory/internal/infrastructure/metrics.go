package infrastructure

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	IncRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method"},
	)

	IncResponseCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_responses_total",
			Help: "Total number of gRPC responses",
		},
		[]string{"method", "status"},
	)

	HistRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 16),
		},
		[]string{"method", "status"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(IncRequestCounter)
	prometheus.MustRegister(IncResponseCounter)
	prometheus.MustRegister(HistRequestDuration)
}
