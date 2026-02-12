package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kento/driver/backend/pkg/jwt"
)

func makeAuthContext(role string) context.Context {
	claims := &jwt.Claims{
		UserID:     "user-1",
		EmployeeID: "emp001",
		Role:       role,
		TokenType:  "access",
	}
	return context.WithValue(context.Background(), ClaimsKey, claims)
}

func TestRequireRole_Allowed(t *testing.T) {
	mw := RequireRole("admin", "dispatcher")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []string{"admin", "dispatcher"}
	for _, role := range tests {
		t.Run(role, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(makeAuthContext(role))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("role %q: status = %d, want %d", role, rec.Code, http.StatusOK)
			}
		})
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	mw := RequireRole("admin")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	tests := []string{"dispatcher", "viewer", "driver"}
	for _, role := range tests {
		t.Run(role, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(makeAuthContext(role))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("role %q: status = %d, want %d", role, rec.Code, http.StatusForbidden)
			}
		})
	}
}

func TestRequireRole_NoClaims(t *testing.T) {
	mw := RequireRole("admin")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRole_ErrorResponseFormat(t *testing.T) {
	mw := RequireRole("admin")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(makeAuthContext("driver"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Error.Code != "FORBIDDEN" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "FORBIDDEN")
	}
}

func TestRequireRole_MultipleRolesDriverAllowed(t *testing.T) {
	mw := RequireRole("admin", "dispatcher", "driver")
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(makeAuthContext("driver"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestJWTAuth_IntegrationWithRBAC(t *testing.T) {
	secret := "integration-test-secret"
	token, _ := jwt.GenerateAccessToken(secret, 15*time.Minute, "user-1", "emp001", "driver")

	// Chain: JWTAuth → RequireRole("admin") → handler
	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for driver accessing admin route")
	})

	handler := JWTAuth(secret)(RequireRole("admin")(innerHandler))

	req := httptest.NewRequest("GET", "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
