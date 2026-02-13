package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseIntParam_Default(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	v, ok := parseIntParam(w, r, "limit", 50)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v != 50 {
		t.Errorf("got %d, want 50", v)
	}
}

func TestParseIntParam_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=25", nil)
	w := httptest.NewRecorder()

	v, ok := parseIntParam(w, r, "limit", 50)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v != 25 {
		t.Errorf("got %d, want 25", v)
	}
}

func TestParseIntParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?limit=abc", nil)
	w := httptest.NewRecorder()

	_, ok := parseIntParam(w, r, "limit", 50)
	if ok {
		t.Fatal("expected ok=false for invalid int")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error code = %q, want VALIDATION_ERROR", resp.Error.Code)
	}
}

func TestParseTimeParam_Empty(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	v, ok := parseTimeParam(w, r, "from")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !v.IsZero() {
		t.Errorf("expected zero time, got %v", v)
	}
}

func TestParseTimeParam_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?from=2024-01-01T00:00:00Z", nil)
	w := httptest.NewRecorder()

	v, ok := parseTimeParam(w, r, "from")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if v.Year() != 2024 {
		t.Errorf("year = %d, want 2024", v.Year())
	}
}

func TestParseTimeParam_Invalid(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?from=not-a-time", nil)
	w := httptest.NewRecorder()

	_, ok := parseTimeParam(w, r, "from")
	if ok {
		t.Fatal("expected ok=false for invalid time")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestIsValidPhoneNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"+639123456789", true},
		{"09123456789", true},
		{"1234567", true},
		{"+1", false},
		{"abc", false},
		{"", false},
		{"1234567890123456", false}, // 16 digits
	}
	for _, tt := range tests {
		got := isValidPhoneNumber(tt.input)
		if got != tt.want {
			t.Errorf("isValidPhoneNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsValidGPSCoord(t *testing.T) {
	tests := []struct {
		lat, lng float64
		want     bool
	}{
		{14.5547, 121.0244, true},  // Manila
		{0, 0, true},               // Gulf of Guinea (valid)
		{-90, -180, true},          // extremes
		{90, 180, true},            // extremes
		{91, 0, false},             // lat out of range
		{0, 181, false},            // lng out of range
		{-91, 0, false},
		{0, -181, false},
	}
	for _, tt := range tests {
		got := isValidGPSCoord(tt.lat, tt.lng)
		if got != tt.want {
			t.Errorf("isValidGPSCoord(%v, %v) = %v, want %v", tt.lat, tt.lng, got, tt.want)
		}
	}
}
