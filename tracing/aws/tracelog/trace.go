package tracelog

import (
	"context"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"net/http"
)

const (
	SpanId  = "span_id"
	Id      = "trace_id"
	Sampled = "sampled"
)

func Setup() {
	tracing.Setup(WithTrace)
}

func New(ctx context.Context, req *http.Request) tracing.Tracing {
	return &TracingContext{
		Path:      req.URL.Path,
		ClientIP:  tracing.ClientIP(req),
		RequestID: req.Header.Get(headers.RequestID),
		Service:   tracing.Service,
	}
}

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	value := ctx.Value(tracing.Key())
	if value != nil {
		tc := value.(tracing.Tracing)
		event = tc.WithTrace(ctx, event)
	}
	return event
}

type TracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
	Service   string
}

func (tc *TracingContext) Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		log = log.Str("trace_id", spanCtx.TraceID().String()).Str("span_id", spanCtx.SpanID().String()).
			Str("sampled", cond(spanCtx.IsSampled(), "01", "00"))
	}
	if tc.Service != "" {
		log = log.Dict("serviceContext", zerolog.Dict().Str("service", tc.Service))
	}
	return log.Str("client_ip", tc.ClientIP).
		Str("request_id", tc.RequestID)
}

func (tc *TracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		event = event.Str(Id, spanCtx.TraceID().String()).
			Str(SpanId, spanCtx.SpanID().String()).
			Str(Sampled, cond(span.IsRecording(), "01", "00"))
	}
	if tc.RequestID != "" {
		event = event.Str("request_id", tc.RequestID)
	}
	return event
}

func cond(is bool, trueValue string, falseValue string) string {
	if is {
		return trueValue
	}
	return falseValue
}
