package resty

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/goccha/logging/log"
	"github.com/rs/zerolog"
	"strings"
)

var _logger = &Logger{}

func SetDebug(c *resty.Client, debug bool) *resty.Client {
	if debug {
		return c.OnRequestLog(RequestLogCallback).
			OnResponseLog(ResponseLogCallback).
			SetLogger(_logger).
			SetDebug(debug)
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
		body.RawJSON("body", []byte(req.Body))
	} else {
		body.Str("body", req.Body)
	}
	log.Debug(context.TODO()).Dict("request", body).Send()
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
		body.RawJSON("body", []byte(res.Body))
	} else {
		body.Str("body", res.Body)
	}
	log.Debug(context.TODO()).Dict("response", body).Send()
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
	if !strings.HasPrefix(format, "\n==") {
		log.Debug(context.TODO()).Msgf("RESTY "+format, v...)
	}
}
