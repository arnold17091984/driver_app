package jwt

import (
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID     string `json:"user_id"`
	EmployeeID string `json:"employee_id"`
	Role       string `json:"role"`
	TokenType  string `json:"token_type"`
	jwtgo.RegisteredClaims
}

func GenerateAccessToken(secret string, expiry time.Duration, userID, employeeID, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		EmployeeID: employeeID,
		Role:       role,
		TokenType:  "access",
		RegisteredClaims: jwtgo.RegisteredClaims{
			ExpiresAt: jwtgo.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now()),
		},
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func GenerateRefreshToken(secret string, expiry time.Duration, userID, employeeID, role string) (string, error) {
	claims := Claims{
		UserID:     userID,
		EmployeeID: employeeID,
		Role:       role,
		TokenType:  "refresh",
		RegisteredClaims: jwtgo.RegisteredClaims{
			ExpiresAt: jwtgo.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwtgo.NewNumericDate(time.Now()),
		},
	}
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(tokenString, secret string) (*Claims, error) {
	token, err := jwtgo.ParseWithClaims(tokenString, &Claims{}, func(token *jwtgo.Token) (interface{}, error) {
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
