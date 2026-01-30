package middleware

import (
	"bytes"
	"encoding/json"
	"io"
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
			if r.ContentLength > maxBytes && r.ContentLength != -1 {
				writeRequestTooLarge(w, maxBytes)
				return
			}

			// If content length is unknown (chunked), read up to maxBytes + 1 to enforce limit.
			if r.ContentLength == -1 {
				origBody := r.Body
				limited := io.LimitReader(origBody, maxBytes+1)
				data, err := io.ReadAll(limited)
				_ = origBody.Close()
				if err != nil {
					writeRequestTooLarge(w, maxBytes)
					return
				}
				if int64(len(data)) > maxBytes {
					writeRequestTooLarge(w, maxBytes)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(data))
				r.ContentLength = int64(len(data))
			}

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
		Message: "request body exceeds size limit",
	}
	_ = json.NewEncoder(w).Encode(resp)
}
