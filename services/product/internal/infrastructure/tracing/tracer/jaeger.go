package tracer

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewTracerProvider returns an OpenTelemetry NewTracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// NewTracerProvider will also use a Resource configured with all the information
// about the application.
func NewTracerProvider(url, service string, attrs ...attribute.KeyValue) (*tracesdk.TracerProvider, error) {
	exp, err := jaeger.New(
		jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(url),
		),
	)
	if err != nil {
		return nil, err
	}

	attrs = append(attrs, semconv.ServiceNameKey.String(service))
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource
		tracesdk.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)),
	)

	return tp, nil
}

func SetGlobalTracerProvider(tp *tracesdk.TracerProvider) {
	otel.SetTracerProvider(tp)
}
