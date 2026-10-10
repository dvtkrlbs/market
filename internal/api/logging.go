package api

import (
	"log/slog"
	"net/http"
	"time"
)

// WithRequestLogging logs completed requests, including validation and CORS failures.
// Bodies, headers, and query strings are deliberately excluded.
func WithRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		response := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(response, request)
		status := response.status
		if status == 0 {
			status = http.StatusOK
		}
		level := slog.LevelInfo
		if status >= http.StatusInternalServerError {
			level = slog.LevelError
		} else if status >= http.StatusBadRequest {
			level = slog.LevelWarn
		}
		slog.Log(request.Context(), level, "HTTP request",
			"method", request.Method, "path", request.URL.Path,
			"status", status, "duration", time.Since(started), "bytes", response.bytes)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *loggingResponseWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *loggingResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	// Informational responses do not commit the final status (except upgrades).
	if status >= 100 && status < 200 && status != http.StatusSwitchingProtocols {
		w.ResponseWriter.WriteHeader(status)
		return
	}
	w.ResponseWriter.WriteHeader(status)
	w.status = status
}

func (w *loggingResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.bytes += n
	return n, err
}

func requestErrorHandler(w http.ResponseWriter, request *http.Request, _ error) {
	// Decoder errors can contain submitted values; avoid logging their contents.
	slog.WarnContext(request.Context(), "HTTP request decoding failed", "method", request.Method, "path", request.URL.Path)
	writeAPIError(w, http.StatusBadRequest, "invalid request")
}

func responseErrorHandler(w http.ResponseWriter, request *http.Request, err error) {
	slog.ErrorContext(request.Context(), "HTTP request handler failed", "method", request.Method, "path", request.URL.Path, "error", err)
	writeAPIError(w, http.StatusInternalServerError, "internal server error")
}
