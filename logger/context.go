package logger

import (
	"context"
	"log/slog"
)

func FromContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if reqId, ok := ctx.Value("requestID").(string); ok {
		return logger.With("requestID", reqId)
	}
	return logger
}
