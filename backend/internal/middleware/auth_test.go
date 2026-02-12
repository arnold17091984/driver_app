package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kento/driver/backend/pkg/jwt"
)

const testSecret = "test-secret-key"

func TestJWTAuth_ValidAccessToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			t.Error("expected claims in context")
			return
		}
		if claims.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
		}
		if claims.Role != "admin" {
			t.Errorf("Role = %q, want %q", claims.Role, "admin")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_InvalidBearerFormat(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	tests := []struct {
		name  string
		value string
	}{
		{"no bearer prefix", "just-a-token"},
		{"basic auth", "Basic dXNlcjpwYXNz"},
		{"empty bearer", "Bearer "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.value)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	token, _ := jwt.GenerateAccessToken(testSecret, -1*time.Hour, "user-1", "emp001", "admin")

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	token, _ := jwt.GenerateAccessToken("other-secret", 15*time.Minute, "user-1", "emp001", "admin")

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_RefreshTokenRejected(t *testing.T) {
	token, _ := jwt.GenerateRefreshToken(testSecret, 168*time.Hour, "user-1", "emp001", "admin")

	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for refresh token")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTAuth_ErrorResponseFormat(t *testing.T) {
	handler := JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("error code = %q, want %q", resp.Error.Code, "UNAUTHORIZED")
	}
}

func TestGetClaims_NoClaims(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	claims := GetClaims(req.Context())
	if claims != nil {
		t.Error("expected nil claims for context without claims")
	}
}
