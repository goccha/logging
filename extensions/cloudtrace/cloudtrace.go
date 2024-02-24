package cloudtrace

import (
	"context"
	"math"
	"os"

	gexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/goccha/envar"
	"github.com/goccha/logging/extensions/tracers"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

func WithServiceName(name string) tracers.Option {
	if name == "" {
		name = envar.String("GAE_SERVICE", "K_SERVICE")
	}
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceName(name))
	}
}

func WithVersion(version string) tracers.Option {
	if version == "" {
		version = os.Getenv("K_REVISION")
	}
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceVersion(version))
	}
}

func Attributes(opt ...tracers.Option) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 4)
	attrs = append(attrs, semconv.CloudProviderGCP)
	if envar.Has("GAE_APPLICATION") {
		attrs = append(attrs, semconv.CloudPlatformGCPAppEngine)
	} else if envar.Has("FUNCTION_TARGET") {
		attrs = append(attrs, semconv.CloudPlatformGCPCloudFunctions)
	} else if (envar.Has("K_REVISION") || envar.Has("K_SERVICE")) && envar.Has("K_CONFIGURATION") {
		attrs = append(attrs, semconv.CloudPlatformGCPCloudRun)
	} else if envar.Has("GKE_CLUSTER_NAME") {
		attrs = append(attrs, semconv.CloudPlatformGCPKubernetesEngine)
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
	if len(attrs) > 0 {
		opts = append(opts, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, attrs...)))
	}
	if projectID := envar.String("GOOGLE_CLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT", "GCP_PROJECT", "GCLOUD_PROJECT"); projectID != "" {
		exporter, err := gexporter.New(gexporter.WithProjectID(projectID))
		if err != nil {
			return nil, err
		}
		rsc, err := resource.New(ctx,
			// Use the GCP resource detector to detect information about the GCP platform
			resource.WithDetectors(gcp.NewDetector()),
			// Keep the default detectors
			resource.WithTelemetrySDK(),
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithResource(rsc), sdktrace.WithBatcher(exporter))
	} else {
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
	}
	return opts, nil
}
