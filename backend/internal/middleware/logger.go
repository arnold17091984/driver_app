package middleware

import (
	"log"
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		reqID := chiMiddleware.GetReqID(r.Context())
		log.Printf("[%s] %s %s %d %s remote=%s ua=%s",
			reqID, r.Method, r.URL.Path, wrapped.status,
			time.Since(start), r.RemoteAddr, r.UserAgent())
	})
}
