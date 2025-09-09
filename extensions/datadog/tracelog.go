package datadog

import (
	"context"
	"net/http"
	"strconv"

	"github.com/goccha/logging/tracing"
	"github.com/goccha/logging/tracing/tracelog"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanId     = "dd.span_id"
	TraceId    = "dd.trace_id"
	ServiceKey = "dd.service"
	EnvKey     = "dd.env"
	VersionKey = "dd.version"
)

var _env, _version string

func Env(env string) tracelog.Option {
	return func(_ *tracelog.Config) {
		_env = env
	}
}

func Version(version string) tracelog.Option {
	return func(_ *tracelog.Config) {
		_version = version
	}
}

func Setup(opt ...tracelog.Option) {
	opt = append(opt, tracelog.WithNewFunc(New()))
	tracelog.Setup(opt...)
}

func New() func(ctx context.Context, req *http.Request) tracing.Tracing {
	return func(ctx context.Context, req *http.Request) tracing.Tracing {
		return &TracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: tracelog.TraceConfig().GetRequestId(ctx, req),
			Service:   tracing.Service(),
			Env:       _env,
			Version:   _version,
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
	Env       string
	Version   string
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
			event = event.Str(ServiceKey, tc.Service)
		}
		if tc.Env != "" {
			event = event.Str(EnvKey, tc.Env)
		}
		if tc.Version != "" {
			event = event.Str(VersionKey, tc.Version)
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
