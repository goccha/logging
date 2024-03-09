package ginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/goccha/logging/log"
	"github.com/goccha/logging/tracing"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

func dumpHeaders(c *gin.Context) {
	logger := log.Debug(c.Request.Context())
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			logger.Str(k, v[0])
		}
	}
	logger.Send()
}

func TraceRequest(tracer trace.Tracer, dump bool, f tracing.NewFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if dump {
			dumpHeaders(c)
		}
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(semconv.HTTPURL(c.Request.URL.String()))
		c.Request = c.Request.WithContext(tracing.With(ctx, c.Request, f))
		if dump {
			log.Dump(c.Request.Context(), log.Debug(c.Request.Context())).Send()
		}
		c.Next()
	}
}
