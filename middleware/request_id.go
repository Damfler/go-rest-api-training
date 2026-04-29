package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

func RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := generateId()
		ctx := context.WithValue(r.Context(), "requestID", id)
		w.Header().Set("X-Request-ID", id)
		next(w, r.WithContext(ctx))
	}
}

func generateId() string {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}
