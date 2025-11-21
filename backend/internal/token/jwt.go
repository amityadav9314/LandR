package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secretKey []byte
}

func NewManager(secret string) *Manager {
	return &Manager{secretKey: []byte(secret)}
}

func (m *Manager) NewJWT(userID string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(duration).Unix(),
	})

	return token.SignedString(m.secretKey)
}

func (m *Manager) Verify(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("invalid subject claim")
		}
		return sub, nil
	}

	return "", fmt.Errorf("invalid token")
}
