package tracing

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

var serviceName string

func Service() string {
	return serviceName
}

type contextKey struct{}

var tracingKey = contextKey{}

func Key() interface{} {
	return tracingKey
}

type NewFunc func(ctx context.Context, req *http.Request) Tracing

type Tracing interface {
	WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event
	Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event
}

func With(ctx context.Context, req *http.Request, f NewFunc) context.Context {
	return context.WithValue(ctx, tracingKey, f(ctx, req))
}

func getHeaderValue(req *http.Request, key string) (string, bool) {
	val := req.Header.Get(key)
	if val == "" {
		return "", false
	}
	return strings.TrimSpace(strings.Split(val, ",")[0]), true
}

type Option func()

func ServiceName(name string) Option {
	return func() {
		serviceName = name
	}
}

func TraceOption(f1 TraceFunc, f ...TraceFunc) Option {
	return func() {
		traceFunc = append([]TraceFunc{f1}, f...)
	}
}

func ClientIP(req *http.Request) string {
	for _, key := range _ipHeaders {
		if val, ok := key(req); ok {
			return val
		}
	}
	return ""
}

type TraceFunc func(ctx context.Context, event *zerolog.Event) *zerolog.Event

var traceFunc []TraceFunc

func Setup(opt ...Option) {
	for _, o := range opt {
		o()
	}
}

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	for _, tf := range traceFunc {
		event = tf(ctx, event)
	}
	return event
}
