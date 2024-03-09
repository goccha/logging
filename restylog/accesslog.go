package restylog

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/goccha/http-constants/pkg/headers"
	"github.com/goccha/logging/log"
	"github.com/rs/zerolog"
)

type Logging func(event *zerolog.Event) *zerolog.Event

func WriteLog(ctx context.Context, name string, req *resty.Request, res *resty.Response, f ...Logging) {
	var latency time.Duration
	ua := req.Header.Get(headers.UserAgent)
	var ev *zerolog.Event
	var status int
	if res != nil {
		status = res.StatusCode()
		latency = res.Time()
		ev = log.Info(ctx)
	} else {
		latency = time.Since(req.Time)
		ev = log.Notice(ctx)
	}
	ev.Str("client", name).Dict("httpClient", zerolog.Dict().
		Int("status", status).Str("userAgent", ua).
		Str("requestMethod", req.Method).Str("protocol", req.RawRequest.URL.Scheme).
		Str("requestHost", req.RawRequest.URL.Host).Str("requestPath", req.RawRequest.URL.Path).
		Str("latency", fmt.Sprintf("%vs", latency.Seconds())))
	if len(f) > 0 {
		ev = f[0](ev)
	}
	ev.Send()
}
