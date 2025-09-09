package tracelog

import (
	"context"
	"fmt"

	"github.com/goccha/envar"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	// Deprecated: use cloudtrace.Operation instead.
	Operation = "logging.googleapis.com/operation"
	// Deprecated: use cloudtrace.SpanId instead.
	SpanId = "logging.googleapis.com/spanId"
	// Deprecated: use cloudtrace.Id instead
	Id = "logging.googleapis.com/trace"
	// Deprecated: use cloudtrace.Sampled instead.
	Sampled = "logging.googleapis.com/trace_sampled"
)

// projectID is the Google Cloud Project ID, obtained from the environment variable.
// It checks "GCP_PROJECT" first, then "GOOGLE_CLOUD_PROJECT".
// If neither is set, projectID will be an empty string.
var projectID = envar.String("GCP_PROJECT", "GOOGLE_CLOUD_PROJECT")

// TracingContext holds tracing information for logging.
// Deprecated: Use cloudtrace.TracingContext instead.
type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
	Producer  string
}

// Dump adds tracing information to the log event.
// Deprecated: Use cloudtrace.TracingContext.Dump instead.
func (tc *TracingContext) Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		log = log.Str("trace_id", spanCtx.TraceID().String()).Str("span_id", spanCtx.SpanID().String()).
			Bool("sampled", spanCtx.IsSampled())
	}
	if tc.Service != "" {
		log = log.Dict("serviceContext", zerolog.Dict().Str("service", tc.Service))
	}
	return log.Str("client_ip", tc.ClientIP).
		Str("request_id", tc.RequestID)
}

// WithTrace adds tracing information to the log event.
// Deprecated: Use cloudtrace.TracingContext.WithTrace instead.
func (tc *TracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if projectID != "" {
		span := trace.SpanFromContext(ctx)
		if span != nil {
			spanCtx := span.SpanContext()
			event = event.Str(Id, fmt.Sprintf("project/%s/traces/%s", projectID, spanCtx.TraceID().String())).
				Str(SpanId, spanCtx.SpanID().String()).
				Bool(Sampled, spanCtx.IsSampled())
		}
	}
	if tc.RequestID != "" {
		event = event.Object(Operation, LogEntryOperation{
			Id:       tc.RequestID,
			Producer: tc.Producer,
			First:    nil,
			Last:     nil,
		})
	}
	return event
}

// LogEntryOperation represents the operation field in a log entry.
// Deprecated: Use cloudtrace.LogEntryOperation instead.
type LogEntryOperation struct {
	Id       string `json:"id"`
	Producer string `json:"producer,omitempty"`
	First    *bool  `json:"first,omitempty"`
	Last     *bool  `json:"last,omitempty"`
}

// MarshalZerologObject implements the zerolog.LogObjectMarshaler interface.
// Deprecated: Use cloudtrace.LogEntryOperation.MarshalZerologObject instead.
func (o LogEntryOperation) MarshalZerologObject(e *zerolog.Event) {
	if o.Id != "" {
		e = e.Str("id", o.Id)
	}
	if o.Producer != "" {
		e = e.Str("producer", o.Producer)
	}
	if o.First != nil {
		e = e.Bool("first", *o.First)
	}
	if o.Last != nil {
		_ = e.Bool("last", *o.Last)
	}
}
