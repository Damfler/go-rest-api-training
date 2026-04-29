package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(debug bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := newResponseWriter(w)
			next(rw, r)

			if debug {
				reqID, _ := r.Context().Value("requestID").(string)
				fmt.Printf("[%s] %s %s %d %v\n", reqID, r.Method, r.URL.Path, rw.statusCode, time.Since(start))
			}
		}
	}
}
