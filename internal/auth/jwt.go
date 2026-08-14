package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/Gui97p/lia-server/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	GroupIDs     []string   `json:"group_ids"`
	TrustLevel   TrustLevel `json:"trust_level"`
	TokenVersion int        `json:"token_version"`
	jwt.RegisteredClaims
}

func GenerateToken(secret []byte, u *users.User) (string, error) {
	claims := Claims{
		UserID:       u.ID.String(),
		Username:     u.Username,
		GroupIDs:     make([]string, 0),
		TrustLevel:   Authenticated,
		TokenVersion: u.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(1, 0, 0)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(secret []byte, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		return claims, nil
	}

	return nil, errors.New("unknown claims type")
}
