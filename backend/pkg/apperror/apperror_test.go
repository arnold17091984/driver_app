package apperror

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := New(400, "BAD_REQUEST", "something went wrong")
	if err.Error() != "something went wrong" {
		t.Errorf("Error() = %q, want %q", err.Error(), "something went wrong")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		status int
		code   string
	}{
		{"unauthorized", ErrUnauthorized, http.StatusUnauthorized, "UNAUTHORIZED"},
		{"forbidden", ErrForbidden, http.StatusForbidden, "FORBIDDEN"},
		{"not found", ErrNotFound, http.StatusNotFound, "NOT_FOUND"},
		{"bad request", ErrBadRequest, http.StatusBadRequest, "BAD_REQUEST"},
		{"conflict", ErrConflict, http.StatusConflict, "CONFLICT"},
		{"internal", ErrInternal, http.StatusInternalServerError, "INTERNAL_ERROR"},
		{"invalid credentials", ErrInvalidCredentials, http.StatusUnauthorized, "INVALID_CREDENTIALS"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Status != tc.status {
				t.Errorf("Status = %d, want %d", tc.err.Status, tc.status)
			}
			if tc.err.Code != tc.code {
				t.Errorf("Code = %q, want %q", tc.err.Code, tc.code)
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, ErrNotFound)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Error.Code != "NOT_FOUND" {
		t.Errorf("code = %q, want %q", resp.Error.Code, "NOT_FOUND")
	}
}

func TestWriteSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"message": "ok"}
	WriteSuccess(rec, data)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
}

func TestWriteCreated(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"id": "123"}
	WriteCreated(rec, data)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestWriteErrorMsg(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteErrorMsg(rec, 422, "VALIDATION_ERROR", "field required")

	if rec.Code != 422 {
		t.Errorf("status = %d, want %d", rec.Code, 422)
	}

	var resp ErrorResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("code = %q, want %q", resp.Error.Code, "VALIDATION_ERROR")
	}
	if resp.Error.Message != "field required" {
		t.Errorf("message = %q, want %q", resp.Error.Message, "field required")
	}
}
