package log

import (
	"context"
	"github.com/goccha/logging/tracing"
	"github.com/rs/zerolog"
)

func Dump(ctx context.Context, log *zerolog.Event) *zerolog.Event {
	value := ctx.Value(tracing.Key)
	if value == nil {
		return log
	}
	tc := value.(tracing.Tracing)
	return tc.Dump(ctx, log)
}
