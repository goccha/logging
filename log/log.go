package log

import (
	"context"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"time"
)

func init() {
	SetGlobalOut(os.Stdout)
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = "debug"
	}
	if len(level) > 0 {
		switch level {
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

var errorLogger = zerolog.New(os.Stderr).With().Caller().Timestamp().Logger()

func SetGlobalErr(w io.Writer) {
	errorLogger = zerolog.New(w).With().Caller().Timestamp().Logger()
}

func Default() *zerolog.Event {
	return log.Info().Str("severity", "DEFAULT")
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
	if skip != nil && len(skip) > 0 {
		skipCount := zerolog.CallerSkipFrameCount + skip[0]
		logger = zerolog.New(os.Stderr).With().CallerWithSkipFrameCount(skipCount).Timestamp().Logger()
	}
	return logger
}
