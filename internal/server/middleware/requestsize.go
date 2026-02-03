package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RequestSize limits the maximum request body size.
// Set maxBytes <= 0 to disable the limit.
func RequestSize(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For known Content-Length, reject early if it exceeds the limit
			if r.ContentLength > maxBytes && r.ContentLength != -1 {
				writeRequestTooLarge(w, maxBytes)
				return
			}

			// Stream-friendly size enforcement: handlers will receive MaxBytesError
			// if they read beyond the limit. We avoid buffering responses to keep
			// memory usage predictable for large/streaming responses.
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func writeRequestTooLarge(w http.ResponseWriter, maxBytes int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusRequestEntityTooLarge)

	resp := errorResponse{
		Error:   "request_too_large",
		Message: fmt.Sprintf("request body exceeds size limit (%d bytes)", maxBytes),
	}
	_ = json.NewEncoder(w).Encode(resp)
}
