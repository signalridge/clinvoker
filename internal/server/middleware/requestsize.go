package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync/atomic"
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

			// For chunked encoding or when Content-Length is within limit, use streaming
			// size checking that doesn't buffer the entire body in memory.
			var exceeded atomic.Bool
			limitedBody := &limitedReader{
				r:        r.Body,
				maxBytes: maxBytes,
				exceeded: &exceeded,
			}
			r.Body = limitedBody

			// Wrap response writer to buffer response until we drain the body
			sw := &sizeCheckingWriter{
				ResponseWriter: w,
				exceeded:       &exceeded,
				maxBytes:       maxBytes,
				bufferedBody:   &bytes.Buffer{},
			}

			// Run handler
			next.ServeHTTP(sw, r)

			// Drain remaining body after handler completes.
			// This ensures large payloads are detected even if handler doesn't read body.
			drainBody(limitedBody, &exceeded)

			// Now flush the response (or 413 if exceeded during drain)
			sw.flushResponse()
		})
	}
}

// limitedReader wraps a reader and tracks if the read exceeded the limit.
// Unlike io.LimitReader, it sets a flag when exceeded instead of silently truncating.
type limitedReader struct {
	r         io.ReadCloser
	maxBytes  int64
	bytesRead int64
	exceeded  *atomic.Bool
}

func (lr *limitedReader) Read(p []byte) (n int, err error) {
	if lr.bytesRead > lr.maxBytes {
		lr.exceeded.Store(true)
		return 0, &http.MaxBytesError{Limit: lr.maxBytes}
	}

	n, err = lr.r.Read(p)
	lr.bytesRead += int64(n)

	if lr.bytesRead > lr.maxBytes {
		lr.exceeded.Store(true)
		return n, &http.MaxBytesError{Limit: lr.maxBytes}
	}

	return n, err
}

func (lr *limitedReader) Close() error {
	return lr.r.Close()
}

// drainBody reads and discards remaining body content.
// This ensures large payloads are detected even if handler doesn't read body.
// If the limit is exceeded during draining, the exceeded flag is set.
func drainBody(lr *limitedReader, exceeded *atomic.Bool) {
	if exceeded.Load() {
		// Already exceeded, just close
		_ = lr.Close()
		return
	}

	// Drain remaining body up to the limit
	buf := make([]byte, 8192)
	for {
		_, err := lr.Read(buf)
		if err != nil {
			break // EOF or error (including limit exceeded)
		}
	}
	_ = lr.Close()
}

// sizeCheckingWriter buffers the response until body drain completes.
// This allows us to return 413 even if the handler wrote a 200 response
// but the body exceeded the limit during post-handler drain.
type sizeCheckingWriter struct {
	http.ResponseWriter
	exceeded       *atomic.Bool
	maxBytes       int64
	bufferedStatus int
	bufferedHeader http.Header
	bufferedBody   *bytes.Buffer
	flushed        bool
}

func (sw *sizeCheckingWriter) Header() http.Header {
	// Return buffered headers so handler can set them
	if sw.bufferedHeader == nil {
		sw.bufferedHeader = make(http.Header)
	}
	return sw.bufferedHeader
}

func (sw *sizeCheckingWriter) WriteHeader(statusCode int) {
	if sw.bufferedStatus != 0 {
		return // Already captured status
	}
	sw.bufferedStatus = statusCode
}

func (sw *sizeCheckingWriter) Write(b []byte) (int, error) {
	if sw.bufferedStatus == 0 {
		sw.bufferedStatus = http.StatusOK
	}
	// If already flushed (streaming mode), write directly
	if sw.flushed {
		if sw.exceeded.Load() {
			return len(b), nil // Discard if exceeded
		}
		return sw.ResponseWriter.Write(b)
	}
	return sw.bufferedBody.Write(b)
}

// flushResponse writes the buffered response to the underlying writer,
// or writes 413 if the body exceeded the limit during drain.
func (sw *sizeCheckingWriter) flushResponse() {
	if sw.flushed {
		return
	}
	sw.flushed = true

	// If body exceeded limit (including during drain), return 413
	if sw.exceeded.Load() {
		writeRequestTooLarge(sw.ResponseWriter, sw.maxBytes)
		return
	}

	// Copy buffered headers to underlying writer
	for key, values := range sw.bufferedHeader {
		for _, value := range values {
			sw.ResponseWriter.Header().Add(key, value)
		}
	}

	// Write status (default to 200 if not set)
	status := sw.bufferedStatus
	if status == 0 {
		status = http.StatusOK
	}
	sw.ResponseWriter.WriteHeader(status)

	// Write buffered body
	_, _ = sw.ResponseWriter.Write(sw.bufferedBody.Bytes())
}

// Flush forwards to the underlying ResponseWriter if it implements http.Flusher.
// This is essential for SSE (Server-Sent Events) streaming.
// When Flush is called, we flush any buffered content and switch to direct mode.
func (sw *sizeCheckingWriter) Flush() {
	// For streaming, flush buffered headers/status even if body is empty.
	if !sw.flushed {
		sw.flushResponse()
	}

	if flusher, ok := sw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack forwards to the underlying ResponseWriter if it implements http.Hijacker.
// This is essential for WebSocket upgrades.
// Note: After hijacking, the connection is taken over and buffering is bypassed.
func (sw *sizeCheckingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := sw.ResponseWriter.(http.Hijacker); ok {
		// Flush any buffered content before hijacking
		if !sw.flushed && sw.bufferedBody.Len() > 0 {
			sw.flushResponse()
		}
		sw.flushed = true
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
}

// Push forwards to the underlying ResponseWriter if it implements http.Pusher.
// This is used for HTTP/2 server push.
func (sw *sizeCheckingWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := sw.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

// ReadFrom forwards to the underlying ResponseWriter if it implements io.ReaderFrom.
// This is an optimization for copying data.
func (sw *sizeCheckingWriter) ReadFrom(r io.Reader) (int64, error) {
	// If already flushed (streaming mode), forward directly to the underlying writer.
	if sw.flushed {
		if sw.exceeded.Load() {
			return io.Copy(io.Discard, r)
		}
		if rf, ok := sw.ResponseWriter.(io.ReaderFrom); ok {
			return rf.ReadFrom(r)
		}
		return io.Copy(sw.ResponseWriter, r)
	}

	// Otherwise buffer to internal buffer.
	if sw.bufferedBody == nil {
		sw.bufferedBody = &bytes.Buffer{}
	}
	return io.Copy(sw.bufferedBody, r)
}

// Unwrap returns the underlying ResponseWriter.
// This allows middleware to access the original writer if needed.
func (sw *sizeCheckingWriter) Unwrap() http.ResponseWriter {
	return sw.ResponseWriter
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
