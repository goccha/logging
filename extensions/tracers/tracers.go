package tracers

import (
	"context"
	"math"

	"github.com/goccha/envar"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

func NewTracerProvider(ctx context.Context, propagator propagation.TextMapPropagator, opts ...sdktrace.TracerProviderOption) (*sdktrace.TracerProvider, error) {
	if envar.Bool("TRACING_ENABLE") {
		tp := sdktrace.NewTracerProvider(opts...)
		otel.SetTracerProvider(tp)
		if propagator == nil {
			propagator = TextMapPropagator()
		}
		otel.SetTextMapPropagator(propagator)
		return tp, nil
	}
	return nil, nil
}

func TextMapPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

type Option func(attrs []attribute.KeyValue) []attribute.KeyValue

func WithServiceName(name string) Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceName(name))
	}
}

func WithVersion(version string) Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceVersion(version))
	}
}

func TracerProviderOptions(ctx context.Context, attrs ...attribute.KeyValue) ([]sdktrace.TracerProviderOption, error) {
	opts := make([]sdktrace.TracerProviderOption, 0, 4)
	fraction := envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
	sampler := sdktrace.AlwaysSample()
	if !math.IsNaN(fraction) {
		sampler = sdktrace.TraceIDRatioBased(fraction)
	}
	opts = append(opts, sdktrace.WithSampler(sampler))
	if len(attrs) > 0 {
		opts = append(opts, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)))
	}
	endpoint := envar.Get("OTEL_EXPORTER_OTLP_ENDPOINT").String("0.0.0.0:4317")
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithDialOption(grpc.WithBlock()))
	if err != nil {
		return nil, err
	}
	rsc, err := resource.New(ctx,
		// Keep the default detectors
		resource.WithTelemetrySDK(),
	)
	if err != nil {
		return nil, err
	}
	opts = append(opts, sdktrace.WithResource(rsc))
	opts = append(opts, sdktrace.WithResource(rsc), sdktrace.WithBatcher(exporter))

	return opts, nil
}
