package log

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	SetGlobalOut(getWriter())
	SetGlobalErr(getErrorWriter())
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "debug"
	}
	if len(level) > 0 {
		switch level {
		case "trace":
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		case "debug":
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		case "info":
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		case "warn":
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		case "error":
			zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		case "fatal":
			zerolog.SetGlobalLevel(zerolog.FatalLevel)
		case "panic":
			zerolog.SetGlobalLevel(zerolog.PanicLevel)
		case "disabled":
			zerolog.SetGlobalLevel(zerolog.Disabled)
		default:
			zerolog.SetGlobalLevel(zerolog.NoLevel)
		}
	} else {
		zerolog.SetGlobalLevel(zerolog.NoLevel)
	}
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

func SetGlobalOut(w io.Writer) {
	log.Logger = zerolog.New(w).With().Timestamp().Logger()
}

var errorLogger zerolog.Logger

func SetGlobalErr(w io.Writer) {
	errorLogger = zerolog.New(w).With().Caller().Timestamp().Logger()
}

func Default(ctx context.Context) *zerolog.Event {
	return tracing.WithTrace(ctx, log.Trace()).Str("severity", "DEFAULT")
}

func Trace(ctx context.Context) *zerolog.Event {
	return tracing.WithTrace(ctx, log.Trace()).Str("severity", "TRACE")
}

func Debug(ctx context.Context) *zerolog.Event {
	return tracing.WithTrace(ctx, log.Debug()).Str("severity", "DEBUG")
}

func Info(ctx context.Context) *zerolog.Event {
	return tracing.WithTrace(ctx, log.Info()).Str("severity", "INFO")
}

func Notice(ctx context.Context) *zerolog.Event {
	return tracing.WithTrace(ctx, log.Info()).Str("severity", "NOTICE")
}

func Warn(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(log.Logger, skip...)
	return tracing.WithTrace(ctx, logger.Warn()).Str("severity", "WARNING")
}

func Error(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return tracing.WithTrace(ctx, logger.Error()).Str("severity", "ERROR")
}

func Fatal(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return tracing.WithTrace(ctx, logger.Error()).Str("severity", "CRITICAL")
}

func Critical(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return tracing.WithTrace(ctx, logger.Error()).Str("severity", "CRITICAL")
}

func Alert(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return tracing.WithTrace(ctx, logger.Error()).Str("severity", "ALERT")
}

func Emergency(ctx context.Context, skip ...int) *zerolog.Event {
	logger := skipLogger(errorLogger, skip...)
	return tracing.WithTrace(ctx, logger.Error()).Str("severity", "EMERGENCY")
}

func skipLogger(logger zerolog.Logger, skip ...int) zerolog.Logger {
	if len(skip) > 0 {
		skipCount := zerolog.CallerSkipFrameCount + skip[0]
		logger = zerolog.New(os.Stderr).With().CallerWithSkipFrameCount(skipCount).Timestamp().Logger()
	}
	return logger
}

type objectKey struct{}

var objKey = objectKey{}

func WithObject(ctx context.Context, obj zerolog.LogObjectMarshaler) context.Context {
	var objs Objects
	if v := ctx.Value(objKey); v == nil {
		objs = make(Objects, 0, 1)
	} else {
		objs = v.(Objects)
	}
	objs = append(objs, obj)
	return context.WithValue(ctx, objKey, objs)
}

func EmbedObject(ctx context.Context, event *zerolog.Event) *zerolog.Event {
	if v := ctx.Value(objKey); v != nil {
		if objs, ok := v.(Objects); ok {
			objs.EmbedObject(event)
		}
	}
	return event
}

type Objects []zerolog.LogObjectMarshaler

func (objs Objects) EmbedObject(event *zerolog.Event) {
	for _, v := range objs {
		event.EmbedObject(v)
	}
}

type Object map[string]any

func (obj Object) MarshalZerologObject(e *zerolog.Event) {
	for k, v := range obj {
		switch v := v.(type) {
		case string:
			e.Str(k, v)
		case int:
			e.Int(k, v)
		case int64:
			e.Int64(k, v)
		case int32:
			e.Int32(k, v)
		case int16:
			e.Int16(k, v)
		case int8:
			e.Int8(k, v)
		case uint:
			e.Uint(k, v)
		case uint64:
			e.Uint64(k, v)
		case uint32:
			e.Uint32(k, v)
		case uint16:
			e.Uint16(k, v)
		case uint8:
			e.Uint8(k, v)
		case float64:
			e.Float64(k, v)
		case float32:
			e.Float32(k, v)
		case bool:
			e.Bool(k, v)
		case time.Time:
			e.Time(k, v)
		case any:
			e.Interface(k, v)
		}
	}
}
