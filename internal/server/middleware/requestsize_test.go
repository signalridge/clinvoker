package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestSize_LimitsBody(t *testing.T) {
	handler := RequestSize(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", strings.NewReader("too-large-body"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rr.Code)
	}
}

func TestRequestSize_AllowsSmallBody(t *testing.T) {
	handler := RequestSize(20)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", strings.NewReader("small"))
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestRequestSize_ChunkedTooLarge(t *testing.T) {
	handler := RequestSize(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", io.NopCloser(bytes.NewReader([]byte("too-large-body"))))
	req.ContentLength = -1
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", rr.Code)
	}
}

func TestRequestSize_ChunkedWithinLimit(t *testing.T) {
	handler := RequestSize(20)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", io.NopCloser(bytes.NewReader([]byte("small"))))
	req.ContentLength = -1
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestRequestSize_FlusherInterface(t *testing.T) {
	var flushed bool

	handler := RequestSize(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// httptest.ResponseRecorder implements Flusher
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
			flushed = true
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !flushed {
		t.Fatal("expected Flusher interface to be available through wrapped ResponseWriter")
	}
}

func TestRequestSize_UnwrapInterface(t *testing.T) {
	handler := RequestSize(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if Unwrap is available
		type unwrapper interface {
			Unwrap() http.ResponseWriter
		}
		if uw, ok := w.(unwrapper); ok {
			underlying := uw.Unwrap()
			if underlying == nil {
				t.Error("Unwrap returned nil")
			}
		} else {
			t.Error("expected Unwrap interface to be available")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
}

func TestRequestSize_DrainDetectsLargeBody(t *testing.T) {
	// Handler that writes response without reading body
	handler := RequestSize(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Deliberately NOT reading the body
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Send a request with a body that exceeds the limit
	req := httptest.NewRequest("POST", "/test", io.NopCloser(bytes.NewReader([]byte("this-body-exceeds-limit"))))
	req.ContentLength = -1 // Chunked encoding
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// The response should be 413 because the drain detected the oversized body
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413 after drain, got %d", rr.Code)
	}
}
