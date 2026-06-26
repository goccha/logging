package xray

import (
	"context"
	"math"
	"strings"

	sampler "github.com/aws-observability/aws-otel-go/samplers/aws/xray"
	"github.com/goccha/envar"
	"github.com/goccha/logging/tracing"
	lambdadetector "go.opentelemetry.io/contrib/detectors/aws/lambda"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// WithLogGroupARNs ロググループARNを設定する
func WithLogGroupARNs(logGroupARNs ...string) tracing.KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logGroupARNs) > 0 {
			attrs = append(attrs, semconv.AWSLogGroupARNsKey.StringSlice(logGroupARNs))
		}
		return attrs
	}
}

// WithLogGroupNames ロググループ名を設定する
func WithLogGroupNames(logGroupNames ...string) tracing.KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logGroupNames) > 0 {
			attrs = append(attrs, semconv.AWSLogGroupNamesKey.StringSlice(logGroupNames))
		}
		return attrs
	}
}

// WithLogStreamARNsKey ログストリームARNを設定する
func WithLogStreamARNsKey(logStreamARNs ...string) tracing.KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logStreamARNs) > 0 {
			attrs = append(attrs, semconv.AWSLogStreamARNsKey.StringSlice(logStreamARNs))
		}
		return attrs
	}
}

// WithLogStreamNames ログストリーム名を設定する
func WithLogStreamNames(logStreamNames ...string) tracing.KeyValueOption {
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		if len(logStreamNames) > 0 {
			attrs = append(attrs, semconv.AWSLogStreamNamesKey.StringSlice(logStreamNames))
		}
		return attrs
	}
}

// WithIDGenerator X-Ray ID ジェネレーターを使用する
func WithIDGenerator() tracing.TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		return sdktrace.WithIDGenerator(xray.NewIDGenerator()), nil
	}
}

// WithSampler X-Ray リモートサンプリングを使用する
// serviceName サービス名
// cloudPlatform "ec2" / "ecs" / "eks" / "lambda", etc
// opt sampler.Option
//
//	sampler.WithEndpoint(endpoint url.URL)
//	sampler.WithSamplingRulesPollingInterval(polingInterval time.Duration)
//	sampler.WithLogger(l logr.Logger)
func WithSampler(serviceName, cloudPlatform string, opt ...sampler.Option) tracing.TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		s, err := sampler.NewRemoteSampler(ctx, serviceName, cloudPlatform, opt...)
		if err != nil {
			return nil, err
		}
		return sdktrace.WithSampler(s), nil
	}
}

// WithResource リソース属性を設定する
// AWS Lambda 環境の場合は、Lambda用のリソース検出器を使用する
// attr 追加の属性
func WithResource(attr ...attribute.KeyValue) tracing.TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		attrs := make([]attribute.KeyValue, 0, 3+len(attr))
		attrs = append(attrs, semconv.CloudProviderAWS)
		region := envar.Get("AWS_REGION,AWS_DEFAULT_REGION").String("ap-northeast-1")
		attrs = append(attrs, semconv.CloudRegion(region))
		attrs = append(attrs, attr...)
		execEnv := envar.String("AWS_EXECUTION_ENV")
		if strings.HasPrefix(execEnv, "AWS_Lambda_") {
			attrs = append(attrs, semconv.CloudPlatformAWSLambda)
			rsc, err := resource.New(ctx,
				resource.WithDetectors(lambdadetector.NewResourceDetector()),
				resource.WithTelemetrySDK())
			if err != nil {
				return nil, err
			}
			if len(attrs) > 0 {
				rsc, err = resource.Merge(rsc, resource.NewWithAttributes(rsc.SchemaURL(), attrs...))
				if err != nil {
					return nil, err
				}
			}
			return sdktrace.WithResource(rsc), nil
		} else if strings.Contains(execEnv, "_ECS_") {
			attrs = append(attrs, semconv.CloudPlatformAWSECS)
			rsc, err := resource.New(ctx,
				resource.WithTelemetrySDK())
			if err != nil {
				return nil, err
			}
			rsc, err = resource.Merge(rsc, resource.NewWithAttributes(rsc.SchemaURL(), attrs...))
			if err != nil {
				return nil, err
			}
			return sdktrace.WithResource(rsc), nil
		}
		return sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)), nil
	}
}

// TracerProviderOptions returns a slice of TracerProviderOption for configuring the OpenTelemetry TracerProvider.
// Deprecated: use tracers.TracerProviderOptions instead.
func TracerProviderOptions(ctx context.Context, attrs ...attribute.KeyValue) ([]sdktrace.TracerProviderOption, error) {
	opts := make([]sdktrace.TracerProviderOption, 0, 4)
	fraction := envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
	s := sdktrace.AlwaysSample()
	if !math.IsNaN(fraction) {
		s = sdktrace.TraceIDRatioBased(fraction)
	}
	opts = append(opts, sdktrace.WithSampler(s))
	opts = append(opts, sdktrace.WithIDGenerator(xray.NewIDGenerator())) // for xray
	if len(attrs) > 0 {
		opts = append(opts, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)))
	}
	endpoint := envar.Get("OTEL_EXPORTER_OTLP_ENDPOINT").String("0.0.0.0:4317")
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(endpoint))
	if err != nil {
		return nil, err
	}
	var rsc *resource.Resource
	execEnv := envar.String("AWS_EXECUTION_ENV")
	if strings.HasPrefix(execEnv, "AWS_Lambda_") {
		detector := lambdadetector.NewResourceDetector()
		rsc, err = detector.Detect(ctx)
	} else {
		rsc, err = resource.New(ctx,
			resource.WithTelemetrySDK(), // Keep the default detectors
		)
	}
	if err != nil {
		return nil, err
	}
	opts = append(opts, sdktrace.WithResource(rsc), sdktrace.WithBatcher(exporter))
	return opts, nil
}

func TextMapPropagator() propagation.TextMapPropagator {
	return xray.Propagator{}
}
