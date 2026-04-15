package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		fmt.Printf("[%s] %s %s %v\n",
			r.Method, r.URL.Path, r.URL.RawQuery, time.Since(start))
	}
}
