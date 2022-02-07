package tracing

import (
	"context"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/http-constants/pkg/headers/forwarded"
	"github.com/rs/zerolog"
	"net"
	"net/http"
	"strings"
)

var Service string

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

func ClientIP(req *http.Request) (clientIP string) {
	if v := req.Header.Get(headers.Forwarded); v != "" {
		clientIP = forwarded.Parse(v).ClientIP()
	} else if clientIP = strings.TrimSpace(strings.Split(req.Header.Get(headers.XForwardedFor), ",")[0]); clientIP == "" {
		clientIP = strings.TrimSpace(req.Header.Get(headers.XRealIp))
	}
	if clientIP != "" {
		return clientIP
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(req.RemoteAddr)); err != nil && ip != "" {
		return ip
	}
	return req.Header.Get(headers.XEnvoyExternalAddress)
}

type TraceFunc func(ctx context.Context, event *zerolog.Event) *zerolog.Event

var traceFunc TraceFunc

func Setup(f TraceFunc) {
	traceFunc = f
}

func WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if traceFunc != nil {
		event = traceFunc(ctx, event)
	}
	return event
}
