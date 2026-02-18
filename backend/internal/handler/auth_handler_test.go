package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestLogin_EmptyBody tests that login rejects an empty request body.
func TestLogin_EmptyBody(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "VALIDATION_ERROR")
	}
}

// TestLogin_MissingPassword tests that login rejects when password is missing.
func TestLogin_MissingPassword(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	body := `{"employee_id":"admin001"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestLogin_MissingEmployeeID tests that login rejects when employee_id is missing.
func TestLogin_MissingEmployeeID(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	body := `{"password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestLogin_InvalidJSON tests that login rejects invalid JSON.
func TestLogin_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestRefresh_InvalidJSON tests that refresh rejects invalid JSON.
func TestRefresh_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader("bad"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

// TestLogout_ReturnsNoContent tests that logout returns 204.
func TestLogout_ReturnsNoContent(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	req := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

// TestMe_NoClaims tests that /me returns 401 without claims in context.
func TestMe_NoClaims(t *testing.T) {
	h := NewAuthHandler(nil, &mockTokenSvc{}, nil)

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestSwaggerUI_Returns200 tests the Swagger UI endpoint.
func TestSwaggerUI_Returns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/docs", nil)
	rec := httptest.NewRecorder()

	SwaggerUI(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

// TestOpenAPISpec_Returns200 tests the OpenAPI spec endpoint.
func TestOpenAPISpec_Returns200(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/docs/openapi.yaml", nil)
	rec := httptest.NewRecorder()

	OpenAPISpec(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Error("response should contain OpenAPI spec")
	}
}
