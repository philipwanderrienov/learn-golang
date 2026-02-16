package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs incoming HTTP requests and response times.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		log.Printf("[%s] %s %s %s", r.Method, r.URL.Path, r.RemoteAddr, time.Since(startTime))
		next.ServeHTTP(w, r)
		elapsedTime := time.Since(startTime)
		log.Printf("[%s] %s completed in %v", r.Method, r.URL.Path, elapsedTime)
	})
}

// RecoveryMiddleware recovers from panics and returns an error response.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC] %s %s: %v", r.Method, r.URL.Path, err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"code":"01","message":"Internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
