package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Logging is a middleware that logs HTTP request details including method, path, status, size, and duration.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		rw := &responseWriter{
			ResponseWriter: w,
			status:         0,
			size:           0,
		}
		
		next.ServeHTTP(rw, r)
		
		duration := time.Since(start)
		
		fmt.Printf("[%s] %s %s %d %d %v\n", 
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			rw.status,
			rw.size,
			duration,
		)
	})
}