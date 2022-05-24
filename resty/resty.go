package resty

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/goccha/logging/log"
	"github.com/rs/zerolog"
	"strings"
)

const (
	Json    = "json"
	Default = "default"
)

var _logger = &Logger{}

func SetFormat(format string) {
	switch format {
	case Json:
		debugFormat = Json
	default:
		debugFormat = Default
	}
}

var debugFormat = Default

func SetDebug(c *resty.Client, debug bool) *resty.Client {
	if debug {
		if debugFormat == Json {
			return c.OnRequestLog(RequestLogCallback).
				OnResponseLog(ResponseLogCallback).
				SetLogger(_logger).
				SetDebug(debug)
		}
		return c.SetDebug(debug)
	}
	return c
}

func RequestLogCallback(req *resty.RequestLog) error {
	body := zerolog.Dict()
	headers := zerolog.Dict()
	for k, v := range req.Header {
		headers.Strs(k, v)
	}
	body.Dict("headers", headers)
	if strings.HasPrefix(req.Body, "{") {
		body.RawJSON("body", []byte(strings.ReplaceAll(req.Body, "\n", "")))
	} else {
		body.Str("body", req.Body)
	}
	log.Debug(context.TODO()).Str("client", "resty").Dict("request", body).Send()
	return nil
}

func ResponseLogCallback(res *resty.ResponseLog) error {
	body := zerolog.Dict()
	headers := zerolog.Dict()
	for k, v := range res.Header {
		headers.Strs(k, v)
	}
	body.Dict("headers", headers)
	if strings.HasPrefix(res.Body, "{") {
		body.RawJSON("body", []byte(strings.ReplaceAll(res.Body, "\n", "")))
	} else {
		body.Str("body", res.Body)
	}
	log.Debug(context.TODO()).Str("client", "resty").Dict("response", body).Send()
	return nil
}

type Logger struct{}

func (l *Logger) Errorf(format string, v ...interface{}) {
	log.Error(context.TODO()).Msgf("RESTY "+format, v...)
}
func (l *Logger) Warnf(format string, v ...interface{}) {
	log.Warn(context.TODO()).Msgf("RESTY "+format, v...)
}
func (l *Logger) Debugf(format string, v ...interface{}) {
	if len(v) > 0 {
		if str, ok := v[0].(string); ok {
			if strings.HasPrefix(str, "\n==") {
				return
			}
		}
	}
	log.Debug(context.TODO()).Msgf("RESTY "+format, v...)
}
