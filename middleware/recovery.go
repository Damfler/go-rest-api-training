package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
)

func Recovery(logger *slog.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					reqID, _ := r.Context().Value("requestID").(string)
					logger.Error("panic recovered",
						"requestID", reqID,
						"method", r.Method,
						"path", r.URL.Path,
						"error", fmt.Sprintf("%v", err),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next(w, r)
		}
	}
}
