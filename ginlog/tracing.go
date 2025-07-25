package ginlog

import (
	"github.com/gin-gonic/gin"
	"github.com/goccha/http-constants/pkg/headers"
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
	logger.Msg("dumpHeaders")
}

func TraceRequest(tracer trace.Tracer, dump bool, f tracing.NewFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if dump {
			dumpHeaders(c)
		}
		ctx := c.Request.Context()
		span := trace.SpanFromContext(ctx)
		span.SetAttributes(semconv.HTTPScheme(c.Request.URL.Scheme))
		span.SetAttributes(semconv.HTTPMethod(c.Request.Method))
		span.SetAttributes(semconv.HTTPTarget(c.Request.URL.Path))
		span.SetAttributes(semconv.HTTPURL(c.Request.URL.String()))
		if l := c.Request.ContentLength; l > 0 {
			span.SetAttributes(semconv.HTTPRequestContentLength(int(l)))
		}
		if ip := tracing.ClientIP(c.Request); ip != "" {
			span.SetAttributes(semconv.HTTPClientIP(ip))
		}
		if ua := c.Request.Header.Get(headers.UserAgent); ua != "" {
			span.SetAttributes(semconv.HTTPUserAgent(ua))
		}
		c.Request = c.Request.WithContext(tracing.With(ctx, c.Request, f))
		if dump {
			log.Dump(ctx, log.Debug(ctx)).Msg("dump")
		}
		c.Next()
		span.SetAttributes(semconv.HTTPStatusCode(c.Writer.Status()))
		if c.Writer.Size() > 0 {
			span.SetAttributes(semconv.HTTPResponseContentLength(c.Writer.Size()))
		}
	}
}
