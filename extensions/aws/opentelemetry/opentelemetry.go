package opentelemetry

import (
	"context"
	"math"

	"github.com/goccha/envar"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
)

func NewTracerProvider(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	if envar.Bool("TRACING_ENABLE") {
		attrs := make([]attribute.KeyValue, 0, 4)
		opts := make([]sdktrace.TracerProviderOption, 0, 4)
		fraction := envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
		sampler := sdktrace.AlwaysSample()
		if !math.IsNaN(fraction) {
			sampler = sdktrace.TraceIDRatioBased(fraction)
		}
		opts = append(opts, sdktrace.WithSampler(sampler))
		opts = append(opts, sdktrace.WithIDGenerator(xray.NewIDGenerator()))
		attrs = append(attrs, semconv.CloudProviderAWS)

		region := envar.Get("AWS_REGION,AWS_DEFAULT_REGION").String("ap-northeast-1")
		attrs = append(attrs, semconv.CloudRegionKey.String(region))

		endpoint := envar.Get("OTEL_EXPORTER_OTLP_ENDPOINT").String("0.0.0.0:4317")
		if exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock())); err != nil {

			return nil, err
		} else {
			opts = append(opts, sdktrace.WithBatcher(exporter))
		}
		// the service name used to display traces in backends
		attrs = append(attrs, semconv.ServiceNameKey.String(serviceName))
		if len(attrs) > 0 {
			res := resource.NewWithAttributes(
				semconv.SchemaURL,
				attrs...,
			)
			opts = append(opts, sdktrace.WithResource(res))
		}
		tp := sdktrace.NewTracerProvider(opts...)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		return tp, nil
	}
	return nil, nil
}
