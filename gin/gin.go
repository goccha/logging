package gin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/log"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"time"
)

func AccessLog(f ...func(e *zerolog.Event) *zerolog.Event) gin.HandlerFunc {
	if len(f) > 0 {
		return JsonLogger(f[0])
	}
	return JsonLogger(nil)
}

func JsonLogger(f func(e *zerolog.Event) *zerolog.Event, notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		// Process request
		c.Next()
		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			// Stop timer
			end := time.Now()
			latency := end.Sub(start)
			ua := c.Request.Header.Get(headers.UserAgent)
			comment := c.Errors.ByType(gin.ErrorTypePrivate).String()
			requestUrl := c.Request.URL.String()
			if c.Request.URL.Scheme == "" {
				scheme := "http"
				if c.Request.TLS != nil {
					scheme = "https"
				}
				requestUrl = fmt.Sprintf("%s://%s%s", scheme, c.Request.Host, requestUrl)
			}
			contentLength := c.Writer.Size()
			e := log.Info(c.Request.Context()).Dict("httpRequest", zerolog.Dict().
				Int("status", c.Writer.Status()).Str("remoteIp", tracing.ClientIP(c.Request)).
				Str("userAgent", ua).Str("latency", fmt.Sprintf("%vs", latency.Seconds())).
				Str("requestMethod", c.Request.Method).Str("requestUrl", requestUrl).
				Str("protocol", c.Request.Proto).Int64("requestSize", c.Request.ContentLength).
				Int("responseSize", contentLength))
			if f != nil {
				f(e).Msg(comment)
			} else {
				e.Msg(comment)
			}
		}
	}
}
