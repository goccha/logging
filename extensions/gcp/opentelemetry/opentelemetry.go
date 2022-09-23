package opentelemetry

import (
	"context"
	"math"
	"os"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/goccha/envar"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func NewTracerProvider(ctx context.Context, serviceName string) (*sdktrace.TracerProvider, error) {
	if envar.Bool("TRACING_ENABLE") {
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		exporter, err := texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			return nil, err
		}

		// Identify your application using resource detection
		res, err := resource.New(ctx,
			// Use the GCP resource detector to detect information about the GCP platform
			resource.WithDetectors(gcp.NewDetector()),
			// Keep the default detectors
			resource.WithTelemetrySDK(),
			// Add your own custom attributes to identify your application
			resource.WithAttributes(
				semconv.ServiceNameKey.String(serviceName),
			),
		)
		if err != nil {
			return nil, err
		}

		// Create trace provider with the exporter.
		//
		// By default it uses AlwaysSample() which samples all traces.
		// In a production environment or high QPS setup please use
		// probabilistic sampling.
		// Example:
		//   tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.0001)), ...)
		fraction := envar.Get("TRACE_ID_RATIO_BASE").Float64(math.NaN())
		sampler := sdktrace.AlwaysSample()
		if !math.IsNaN(fraction) {
			sampler = sdktrace.TraceIDRatioBased(fraction)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sampler),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		return tp, nil
	}
	return nil, nil
}
