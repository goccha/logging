package tracelog

import (
	"context"
	"net/http"
	"strconv"

	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanId            = "dd.span_id"
	TraceId           = "dd.trace_id"
	DatadogServiceKey = "dd.service"
	DatadogEnvKey     = "dd.env"
	DatadogVersionKey = "dd.version"
)

type Option func()

func Env(env string) Option {
	return func() {
		_env = env
	}
}

func Version(version string) Option {
	return func() {
		_version = version
	}
}

func Service(service string) Option {
	return func() {
		tracing.Service = service
	}
}

func Setup(options ...Option) {
	tracing.Setup(WithTrace)
	if len(options) > 0 {
		for _, opt := range options {
			opt()
		}
	}
}

var _env, _version string

func New() func(ctx context.Context, req *http.Request) tracing.Tracing {
	return func(ctx context.Context, req *http.Request) tracing.Tracing {
		return &TracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: req.Header.Get(headers.RequestID),
			Service:   tracing.Service,
		}
	}
}

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	value := ctx.Value(tracing.Key)
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
		event = event.Str(TraceId, convert(spanCtx.TraceID().String())).
			Str(SpanId, convert(spanCtx.SpanID().String()))
		if tc.Service != "" {
			event = event.Str(DatadogServiceKey, tc.Service)
		}
		if _env != "" {
			event = event.Str(DatadogEnvKey, _env)
		}
		if _version != "" {
			event = event.Str(DatadogVersionKey, _version)
		}
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

func convert(id string) string {
	if len(id) < 16 {
		return ""
	}
	if len(id) > 16 {
		id = id[16:]
	}
	intValue, err := strconv.ParseUint(id, 16, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatUint(intValue, 10)
}
