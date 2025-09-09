package tracelog

import (
	"context"
	"net/http"

	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanId  = "span_id"
	Id      = "trace_id"
	Sampled = "sampled"
)

type Config struct {
	RequestIdHeader string
	RequestIdFunc
	tracing.NewFunc
}

type RequestIdFunc func(ctx context.Context, req *http.Request) string

var _config = &Config{}

func (c *Config) GetRequestId(ctx context.Context, req *http.Request) string {
	if c.RequestIdFunc != nil {
		return c.RequestIdFunc(ctx, req)
	}
	if c.RequestIdHeader != "" {
		return req.Header.Get(_config.RequestIdHeader)
	}
	return req.Header.Get(headers.RequestID)
}

func TraceConfig() *Config {
	return _config
}

type Option func(c *Config)

func WithRequestIdHeader(header string) Option {
	return func(c *Config) {
		c.RequestIdHeader = header
	}
}

func WithRequestIdFunc(f RequestIdFunc) Option {
	return func(c *Config) {
		c.RequestIdFunc = f
	}
}

func WithNewFunc(f tracing.NewFunc) Option {
	return func(c *Config) {
		c.NewFunc = f
	}
}

func Setup(opt ...Option) {
	tracing.Setup(tracing.LogOption(WithTrace()))
	if len(opt) > 0 {
		for _, op := range opt {
			op(_config)
		}
	}
}

func New() func(ctx context.Context, req *http.Request) tracing.Tracing {
	if _config.NewFunc != nil {
		return _config.NewFunc
	}
	return func(ctx context.Context, req *http.Request) tracing.Tracing {
		return &TracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: _config.GetRequestId(ctx, req),
			Service:   tracing.Service(),
		}
	}
}

func WithTrace() tracing.LogFunc {
	return func(ctx context.Context, event *zerolog.Event) *zerolog.Event {
		if value := tracing.Value(ctx); value != nil {
			if tc, ok := value.(tracing.Tracing); ok {
				event = tc.WithTrace(ctx, event)
			}
		}
		return event
	}
}

func WithContext(ctx context.Context, req *http.Request, f ...tracing.NewFunc) context.Context {
	var tc tracing.Tracing
	if len(f) > 0 {
		tc = f[0](ctx, req)
	} else {
		tc = New()(ctx, req)
	}
	return tracing.WithContext(ctx, tc)
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
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		event = event.Str(Id, spanCtx.TraceID().String()).
			Str(SpanId, spanCtx.SpanID().String()).
			Str(Sampled, cond(spanCtx.IsSampled(), "01", "00"))
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
