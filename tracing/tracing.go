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

func Value(ctx context.Context) interface{} {
	value := ctx.Value(tracingKey)
	if value != nil {
		return value
	}
	return nil
}

func WithContext(ctx context.Context, tr Tracing) context.Context {
	return context.WithValue(ctx, tracingKey, tr)
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

func LogOption(f1 LogFunc, f ...LogFunc) Option {
	return func() {
		if len(logFunc) == 0 {
			logFunc = append([]LogFunc{f1}, f...)
		} else {
			logFunc = append(logFunc, f1)
			logFunc = append(logFunc, f...)
		}
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

type LogFunc func(ctx context.Context, event *zerolog.Event) *zerolog.Event

var logFunc []LogFunc

func Setup(opt ...Option) {
	for _, o := range opt {
		o()
	}
}

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	for _, tf := range logFunc {
		event = tf(ctx, event)
	}
	return event
}
