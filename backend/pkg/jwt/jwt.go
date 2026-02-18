package jwt

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	jwtgo "github.com/golang-jwt/jwt/v5"
)

const (
	Issuer   = "fleettrack"
	Audience = "fleettrack-api"
)

type Claims struct {
	UserID     string `json:"user_id"`
	EmployeeID string `json:"employee_id"`
	Role       string `json:"role"`
	TokenType  string `json:"token_type"`
	jwtgo.RegisteredClaims
}

func GenerateAccessToken(secret string, expiry time.Duration, userID, employeeID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		EmployeeID: employeeID,
		Role:       role,
		TokenType:  "access",
		RegisteredClaims: jwtgo.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    Issuer,
			Subject:   userID,
			Audience:  jwtgo.ClaimStrings{Audience},
			ExpiresAt: jwtgo.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwtgo.NewNumericDate(now),
			NotBefore: jwtgo.NewNumericDate(now),
		},
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(secret string, expiry time.Duration, userID, employeeID, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:     userID,
		EmployeeID: employeeID,
		Role:       role,
		TokenType:  "refresh",
		RegisteredClaims: jwtgo.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    Issuer,
			Subject:   userID,
			Audience:  jwtgo.ClaimStrings{Audience},
			ExpiresAt: jwtgo.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwtgo.NewNumericDate(now),
			NotBefore: jwtgo.NewNumericDate(now),
		},
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(tokenString, secret string) (*Claims, error) {
	token, err := jwtgo.ParseWithClaims(tokenString, &Claims{}, func(token *jwtgo.Token) (interface{}, error) {
		// Validate algorithm to prevent alg-switching attacks
		if _, ok := token.Method.(*jwtgo.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwtgo.ErrTokenInvalidClaims
	}
	return claims, nil
}
