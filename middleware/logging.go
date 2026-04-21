package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Logging(debug bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next(w, r)

			if debug {
				fmt.Printf("[%s] %s %s query=%s duration=%v\n",
					r.Method, r.URL.Path, r.RemoteAddr,
					r.URL.RawQuery, time.Since(start))
			} else {
				fmt.Printf("[%s] %s %s %v\n",
					r.Method, r.URL.Path, r.URL.RawQuery, time.Since(start))
			}
		}
	}
}
