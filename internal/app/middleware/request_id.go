package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
)

type contextKey string

const requestIDKey contextKey = "request_id"

var reqCounter uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	count := atomic.AddUint64(&reqCounter, 1)
	return "req-" + strconv.FormatUint(count, 10)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
