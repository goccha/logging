package cloudtrace

import (
	"context"
	"fmt"
	"net/http"

	"github.com/goccha/envar"
	"github.com/goccha/logging/tracing"
	"github.com/goccha/logging/tracing/tracelog"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	Operation = "logging.googleapis.com/operation"
	SpanId    = "logging.googleapis.com/spanId"
	Id        = "logging.googleapis.com/trace"
	Sampled   = "logging.googleapis.com/trace_sampled"
)

func init() {
	tracing.Setup(tracing.ServiceName(envar.String("GAE_SERVICE", "K_SERVICE")))
}

var projectID = envar.String("GCP_PROJECT", "GOOGLE_CLOUD_PROJECT")

func Setup(opt ...tracelog.Option) {
	opt = append(opt, tracelog.WithNewFunc(New()))
	tracelog.Setup(opt...)
}

func New() func(ctx context.Context, req *http.Request) tracing.Tracing {
	return func(ctx context.Context, req *http.Request) tracing.Tracing {
		config := tracelog.TraceConfig()
		return &TracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: config.GetRequestId(ctx, req),
			Service:   tracing.Service(),
			Producer:  envar.Get("GOOGLE_TRACE_PRODUCER").String(""),
		}
	}
}

func Context(ctx context.Context) *TracingContext {
	value := ctx.Value(tracing.Key())
	if value != nil {
		if tc, ok := value.(*TracingContext); ok {
			return tc
		}
	}
	return nil
}

type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
	Producer  string
}

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

type LogEntryOperation struct {
	Id       string `json:"id"`
	Producer string `json:"producer,omitempty"`
	First    *bool  `json:"first,omitempty"`
	Last     *bool  `json:"last,omitempty"`
}

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
