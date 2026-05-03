package logger

import (
	"context"
	"log/slog"
)

func FromContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if reqId, ok := ctx.Value("requestId").(string); ok {
		return logger.With("requestId", reqId)
	}
	return logger
}
