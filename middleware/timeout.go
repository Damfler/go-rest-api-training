package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(seconds int) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
			defer cancel()

			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}
