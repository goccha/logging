package xray

import (
	"context"
	"math"
	"strings"

	"github.com/goccha/envar"
	"github.com/goccha/logging/extensions/tracers"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

func WithLogGroupARNs(logGroupARNs ...string) tracers.Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logGroupARNs) > 0 {
			attrs = append(attrs, semconv.AWSLogGroupARNsKey.StringSlice(logGroupARNs))
		}
		return attrs
	}
}

func WithLogGroupNames(logGroupNames ...string) tracers.Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logGroupNames) > 0 {
			attrs = append(attrs, semconv.AWSLogGroupNamesKey.StringSlice(logGroupNames))
		}
		return attrs
	}
}

func WithLogStreamARNsKey(logStreamARNs ...string) tracers.Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logStreamARNs) > 0 {
			attrs = append(attrs, semconv.AWSLogStreamARNsKey.StringSlice(logStreamARNs))
		}
		return attrs
	}
}

func WithLogStreamNames(logStreamNames ...string) tracers.Option {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logStreamNames) > 0 {
			attrs = append(attrs, semconv.AWSLogStreamNamesKey.StringSlice(logStreamNames))
		}
		return attrs
	}
}

func Attributes(opt ...tracers.Option) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 4)
	attrs = append(attrs, semconv.CloudProviderAWS)
	region := envar.Get("AWS_REGION,AWS_DEFAULT_REGION").String("ap-northeast-1")
	attrs = append(attrs, semconv.CloudRegion(region))
	execEnv := envar.String("AWS_EXECUTION_ENV")
	if strings.HasPrefix(execEnv, "AWS_Lambda_") {
		attrs = append(attrs, semconv.CloudPlatformAWSLambda)
	} else if strings.Contains(execEnv, "_ECS_") {
		attrs = append(attrs, semconv.CloudPlatformAWSECS)
	}
	for _, o := range opt {
		attrs = o(attrs)
	}
	return attrs
}

func TracerProviderOptions(ctx context.Context, attrs ...attribute.KeyValue) ([]sdktrace.TracerProviderOption, error) {
	opts := make([]sdktrace.TracerProviderOption, 0, 4)
	fraction := envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
	sampler := sdktrace.AlwaysSample()
	if !math.IsNaN(fraction) {
		sampler = sdktrace.TraceIDRatioBased(fraction)
	}
	opts = append(opts, sdktrace.WithSampler(sampler))
	opts = append(opts, sdktrace.WithIDGenerator(xray.NewIDGenerator())) // for xray
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
	opts = append(opts, sdktrace.WithResource(rsc), sdktrace.WithBatcher(exporter))
	return opts, nil
}

func TextMapPropagator() propagation.TextMapPropagator {
	return xray.Propagator{}
}
