package cloudtrace

import (
	"context"
	"math"

	detectors "github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp"
	exporters "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/goccha/envar"
	"github.com/goccha/logging/tracing"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func WithServiceName(name string) tracing.KeyValueOption {
	if name == "" {
		name = envar.String("GAE_SERVICE", "K_SERVICE")
	}
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceName(name))
	}
}

func WithVersion(version string) tracing.KeyValueOption {
	if version == "" {
		version = envar.String("K_REVISION")
	}
	return func(attrs []attribute.KeyValue) []attribute.KeyValue {
		return append(attrs, semconv.ServiceVersion(version))
	}
}

func ProjectId() string {
	projectId, err := detectors.NewDetector().ProjectID()
	if err != nil {
		return envar.String("GOOGLE_CLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT", "GCP_PROJECT", "GCLOUD_PROJECT")
	}
	return projectId
}

func WithExporter(projectId string) tracing.TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		exporter, err := exporters.New(exporters.WithProjectID(projectId))
		if err != nil {
			return nil, err
		}
		return sdktrace.WithBatcher(exporter), nil
	}
}

func WithResource(attrs ...attribute.KeyValue) tracing.TracerProviderOption {
	return func(ctx context.Context) (sdktrace.TracerProviderOption, error) {
		rsc, err := resource.New(ctx,
			resource.WithDetectors(gcp.NewDetector()),
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
	}
}

// TracerProviderOptions returns a slice of TracerProviderOption for configuring the OpenTelemetry TracerProvider.
// Deprecated: use tracing.TracerProviderOptions instead.
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
		exporter, err := exporters.New(exporters.WithProjectID(projectID))
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
			otlptracegrpc.WithEndpoint(endpoint))
		if err != nil {
			return nil, err
		}
		rsc, err := resource.New(ctx,
			resource.WithTelemetrySDK(), // Keep the default detectors
		)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdktrace.WithResource(rsc), sdktrace.WithBatcher(exporter))
	}
	return opts, nil
}
