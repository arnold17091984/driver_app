package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":          "DENY",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Cache-Control":             "no-store",
		"X-XSS-Protection":         "0",
	}

	for name, expected := range expectedHeaders {
		got := rec.Header().Get(name)
		if got != expected {
			t.Errorf("%s = %q, want %q", name, got, expected)
		}
	}
}

func TestMaxBodySize_UnderLimit(t *testing.T) {
	handler := MaxBodySize(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 100)
		_, err := r.Body.Read(buf)
		if err != nil && err.Error() == "http: request body too large" {
			t.Error("should not get body too large for small request")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMaxBodySize_OverLimit(t *testing.T) {
	handler := MaxBodySize(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 100)
		_, err := r.Body.Read(buf)
		if err == nil {
			t.Error("expected error for oversized body")
		}
	}))

	req := httptest.NewRequest("POST", "/test", strings.NewReader("this is a very long body that exceeds the limit"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
