package gin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/log"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
)

func AccessLog(f ...func(c *gin.Context, e *zerolog.Event) *zerolog.Event) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		// Process request
		c.Next()
		// Stop timer
		end := time.Now()
		latency := end.Sub(start)
		JsonLog(c, func(c *gin.Context, e *zerolog.Event) {
			e.Str("latency", fmt.Sprintf("%vs", latency.Seconds()))
		}, f...)
	}
}

func JsonLog(c *gin.Context, f func(c *gin.Context, e *zerolog.Event), filters ...func(c *gin.Context, e *zerolog.Event) *zerolog.Event) {
	req := c.Request
	ctx := req.Context()
	ua := req.Header.Get(headers.UserAgent)
	requestUrl := req.URL.String()
	if req.URL.Scheme == "" {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		requestUrl = fmt.Sprintf("%s://%s%s", scheme, req.Host, requestUrl)
	}
	e := log.Info(ctx).Dict("httpRequest", zerolog.Dict().
		Int("status", c.Writer.Status()).Str("remoteIp", tracing.ClientIP(req)).
		Str("userAgent", ua).
		Str("requestMethod", req.Method).Str("requestUrl", requestUrl).
		Str("protocol", req.Proto).Int64("requestSize", req.ContentLength).
		Int("responseSize", c.Writer.Size()))
	if f != nil {
		f(c, e)
	}
	event := log.EmbedObject(ctx, e)
	for _, filter := range filters {
		if event = filter(c, event); event == nil {
			return
		}
	}
	if event != nil {
		errMsgs := c.Errors.ByType(gin.ErrorTypePrivate)
		if len(errMsgs) > 0 {
			event.Msg(errMsgs.String())
		} else {
			event.Send()
		}
	}
}
