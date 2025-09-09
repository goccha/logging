package tracing

import (
	"context"
	"math"

	"github.com/goccha/envar"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type TracerBuilder struct {
	Propagator propagation.TextMapPropagator
	Options    []sdktrace.TracerProviderOption
}

func NewTracer() *TracerBuilder {
	return &TracerBuilder{
		Propagator: TextMapPropagator(),
		Options:    make([]sdktrace.TracerProviderOption, 0, 4),
	}
}
func (b *TracerBuilder) WithPropagator(propagator propagation.TextMapPropagator) *TracerBuilder {
	if propagator != nil {
		b.Propagator = propagator
	}
	return b
}
func (b *TracerBuilder) WithOptions(opts ...sdktrace.TracerProviderOption) *TracerBuilder {
	if len(opts) > 0 {
		b.Options = append(b.Options, opts...)
	}
	return b
}
func (b *TracerBuilder) Build(ctx context.Context) (*sdktrace.TracerProvider, error) {
	return NewTracerProvider(ctx, b.Propagator, b.Options...)
}

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

type KeyValueOption func(attrs []attribute.KeyValue) []attribute.KeyValue

func WithServiceName(name string) KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceName(name))
	}
}

func WithVersion(version string) KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceVersion(version))
	}
}

func Attributes(options ...KeyValueOption) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 4)
	for _, option := range options {
		attrs = option(attrs)
	}
	return attrs
}

type TracerProviderOption func(ctx context.Context) (sdktrace.TracerProviderOption, error)

func WithSampler(fractions ...float64) TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		var fraction float64
		if len(fractions) > 0 {
			fraction = fractions[0]
		} else {
			fraction = envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
		}
		sampler := sdktrace.AlwaysSample()
		if !math.IsNaN(fraction) {
			sampler = sdktrace.TraceIDRatioBased(fraction)
		}
		return sdktrace.WithSampler(sampler), nil
	}
}

func WithResource(attrs ...attribute.KeyValue) TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		return sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)), nil
	}
}

func WithGrpcExporter(opts ...otlptracegrpc.Option) TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		endpoint := envar.Get("OTEL_EXPORTER_OTLP_ENDPOINT").String("0.0.0.0:4317")
		options := make([]otlptracegrpc.Option, 0, len(opts)+2)
		options = append(options, otlptracegrpc.WithInsecure(), otlptracegrpc.WithEndpoint(endpoint))
		options = append(options, opts...)
		exporter, err := otlptracegrpc.New(ctx, options...)
		if err != nil {
			return nil, err
		}
		return sdktrace.WithBatcher(exporter), nil
	}
}

func WithHttpExporter(opts ...otlptracehttp.Option) TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		endpoint := envar.Get("OTEL_EXPORTER_OTLP_ENDPOINT").String("0.0.0.0:4317")
		options := make([]otlptracehttp.Option, 0, len(opts)+2)
		options = append(options, otlptracehttp.WithInsecure(), otlptracehttp.WithEndpoint(endpoint))
		options = append(options, opts...)
		exporter, err := otlptracehttp.New(ctx, options...)
		if err != nil {
			return nil, err
		}
		return sdktrace.WithBatcher(exporter), nil
	}
}

func TracerProviderOptions(ctx context.Context, options ...TracerProviderOption) ([]sdktrace.TracerProviderOption, error) {
	opts := make([]sdktrace.TracerProviderOption, 0, len(options))
	for _, option := range options {
		opt, err := option(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, opt)
	}
	return opts, nil
}
