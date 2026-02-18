package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key"

func TestGenerateAccessToken(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := GenerateRefreshToken(testSecret, 168*time.Hour, "user-1", "emp001", "driver")
	if err != nil {
		t.Fatalf("GenerateRefreshToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestParseAccessToken(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")

	claims, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-1")
	}
	if claims.EmployeeID != "emp001" {
		t.Errorf("EmployeeID = %q, want %q", claims.EmployeeID, "emp001")
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want %q", claims.Role, "admin")
	}
	if claims.TokenType != "access" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "access")
	}
}

func TestParseRefreshToken(t *testing.T) {
	token, _ := GenerateRefreshToken(testSecret, 168*time.Hour, "user-2", "drv001", "driver")

	claims, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != "user-2" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-2")
	}
	if claims.TokenType != "refresh" {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, "refresh")
	}
}

func TestParseExpiredToken(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, -1*time.Hour, "user-1", "emp001", "admin")

	_, err := Parse(token, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestParseWrongSecret(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")

	_, err := Parse(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestParseInvalidToken(t *testing.T) {
	_, err := Parse("not-a-valid-token", testSecret)
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
}

func TestAccessAndRefreshTokensDiffer(t *testing.T) {
	access, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	refresh, _ := GenerateRefreshToken(testSecret, 168*time.Hour, "user-1", "emp001", "admin")

	if access == refresh {
		t.Error("access and refresh tokens should be different")
	}

	aClaims, _ := Parse(access, testSecret)
	rClaims, _ := Parse(refresh, testSecret)

	if aClaims.TokenType != "access" {
		t.Errorf("access token type = %q, want %q", aClaims.TokenType, "access")
	}
	if rClaims.TokenType != "refresh" {
		t.Errorf("refresh token type = %q, want %q", rClaims.TokenType, "refresh")
	}
}

func TestTokenExpiryIsSet(t *testing.T) {
	expiry := 30 * time.Minute
	token, _ := GenerateAccessToken(testSecret, expiry, "user-1", "emp001", "admin")

	claims, _ := Parse(token, testSecret)

	if claims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt to be set")
	}
	if claims.IssuedAt == nil {
		t.Fatal("expected IssuedAt to be set")
	}

	diff := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if diff < expiry-time.Second || diff > expiry+time.Second {
		t.Errorf("token expiry duration = %v, want ~%v", diff, expiry)
	}
}

func TestTokenHasJTI(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	claims, _ := Parse(token, testSecret)

	if claims.ID == "" {
		t.Fatal("expected JTI (ID) to be set")
	}
}

func TestTokenHasIssuerAudienceSubject(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	claims, _ := Parse(token, testSecret)

	if claims.Issuer != Issuer {
		t.Errorf("Issuer = %q, want %q", claims.Issuer, Issuer)
	}
	if claims.Subject != "user-1" {
		t.Errorf("Subject = %q, want %q", claims.Subject, "user-1")
	}
	aud := claims.Audience
	if len(aud) != 1 || aud[0] != Audience {
		t.Errorf("Audience = %v, want [%q]", aud, Audience)
	}
}

func TestTokenHasNotBefore(t *testing.T) {
	token, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	claims, _ := Parse(token, testSecret)

	if claims.NotBefore == nil {
		t.Fatal("expected NotBefore to be set")
	}
}

func TestTokenJTIsAreUnique(t *testing.T) {
	t1, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")
	t2, _ := GenerateAccessToken(testSecret, 15*time.Minute, "user-1", "emp001", "admin")

	c1, _ := Parse(t1, testSecret)
	c2, _ := Parse(t2, testSecret)

	if c1.ID == c2.ID {
		t.Error("two tokens should have different JTI values")
	}
}
