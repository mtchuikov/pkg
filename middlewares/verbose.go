package middlewares

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type verboseResponseData struct {
	status int
	size   int
}

type verboseResponseWriter struct {
	http.ResponseWriter
	responseData *verboseResponseData
}

func (r *verboseResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *verboseResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Verbose is an HTTP middleware that logs request and response
// details. It records HTTP method, URL, client IP address,
// user-agent, referer, response status code, duration of request
// processing, and response size. This middleware is useful for
// detailed monitoring and debugging of HTTP requests and their
// handling behavior within the service.
func Verbose(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(rw http.ResponseWriter, req *http.Request) {
			start := time.Now()
			respData := verboseResponseData{
				status: 0,
				size:   0,
			}

			vrw := verboseResponseWriter{
				ResponseWriter: rw,
				responseData:   &respData,
			}

			next.ServeHTTP(&vrw, req)
			duration := time.Since(start)

			logger.Info().
				Str("method", req.Method).
				Str("url", req.URL.String()).
				Str("remote_addr", req.RemoteAddr).
				Str("user_agent", req.UserAgent()).
				Int("status", respData.status).
				Dur("duration", duration).
				Int("size", respData.size).
				Str("referer", req.Referer()).
				Msg("request handled")
		}

		return http.HandlerFunc(fn)
	}
}
