package log

import (
	"context"

	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
)

func Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	value := tracing.Value(ctx)
	if value == nil {
		return log
	}
	if tc, ok := value.(tracing.Tracing); ok {
		return tc.Dump(ctx, log)
	}
	return log
}
