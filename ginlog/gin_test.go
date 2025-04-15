package ginlog

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/log"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const (
	spanId  = "span_id"
	Id      = "trace_id"
	sampled = "sampled"
)

type tracingContext struct {
	Path      string
	ClientIP  string
	RequestID string
}

func (tc tracingContext) WithTrace(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		event = event.Str(Id, spanCtx.TraceID().String()).
			Str(spanId, spanCtx.SpanID().String()).
			Str(sampled, "01")
	}
	if tc.RequestID != "" {
		event = event.Str("request_id", tc.RequestID)
	}
	return event
}
func (tc tracingContext) Dump(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		spanCtx := span.SpanContext()
		event = event.Str("trace_id", spanCtx.TraceID().String()).Str("span_id", spanCtx.SpanID().String()).
			Bool("sampled", spanCtx.IsSampled())
	}
	return event.Str("client_ip", tc.ClientIP).
		Str("request_id", tc.RequestID)
}

func TestJsonLogger(t *testing.T) {
	tracing.Setup(tracing.TraceOption(func(ctx context.Context, event *zerolog.Event) *zerolog.Event {
		value := ctx.Value(tracing.Key())
		if value != nil {
			tc := value.(tracing.Tracing)
			event = tc.WithTrace(ctx, event)
		}
		return event
	}))
	reqBody := bytes.NewBufferString("request body")
	req := httptest.NewRequest(http.MethodGet, "http://dummy.url.com/user", reqBody)
	req.Header.Add("User-Agent", "Test/0.0.1")

	buf := &bytes.Buffer{}
	log.SetGlobalOut(buf)

	var tracer = otel.Tracer("github.com/goccha-tracer")
	router := gin.New()
	router.Use(TraceRequest(tracer, true, func(ctx context.Context, req *http.Request) tracing.Tracing {
		return &tracingContext{
			Path:      req.URL.Path,
			ClientIP:  tracing.ClientIP(req),
			RequestID: req.Header.Get(headers.RequestID),
		}
	}), AccessLog()).
		GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	w := PerformRequest(router, "GET", "/test")
	//fmt.Printf("%s\n", buf.String())
	for _, str := range strings.Split(buf.String(), "\n") {
		if len(str) == 0 {
			continue
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal([]byte(str), &m); err != nil {
			t.Error(err)
			return
		}
		if v, ok := m[Id]; ok {
			assert.Equal(t, "00000000000000000000000000000000", v)
		}
		if v, ok := m[spanId]; ok {
			assert.Equal(t, "0000000000000000", v)
		}
	}
	assert.Equal(t, http.StatusOK, w.Code)
}

type header struct {
	Key   string
	Value string
}

func PerformRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
