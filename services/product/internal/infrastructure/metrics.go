package infrastructure

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
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
			Buckets: []float64{.1, .3, .5, .7, 1, 3, 5, 7, 10},
		},
		[]string{"method", "status"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(IncRequestCounter)
	prometheus.MustRegister(IncResponseCounter)
	prometheus.MustRegister(HistRequestDuration)
}

func RunMetricsServer(addr string) {
	InitMetrics()

	mux := http.NewServeMux()

	mux.Handle("/metrics", promhttp.Handler())

	fmt.Printf("starting metrics server on addr %s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("failed to start metrics server: %v\n", err)
	}
}
