package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"manga-reader/internal/apperror"
	"manga-reader/internal/response"
	"net/http"
	"strings"
	"time"
)

var JWTSecret []byte

func SetJWTSecret(secret string) {
	JWTSecret = []byte(secret)
}

func GenerateToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return JWTSecret, nil
	})
	if err != nil {
		return 0, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if uid, ok := claims["user_id"].(float64); ok {
			return int64(uid), nil
		}
		return 0, errors.New("user_id not found in token")
	}
	return 0, errors.New("invalid token")
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			err := apperror.NewUnauthorizedError("Отсутствует заголовок Authorization", nil)
			response.Error(w, nil, err)
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			err := apperror.NewUnauthorizedError("Неверный формат заголовка Authorization", nil)
			response.Error(w, nil, err)
			return
		}
		userID, err := ParseToken(parts[1])
		if err != nil {
			err := apperror.NewUnauthorizedError("Неверный токен: "+err.Error(), err)
			response.Error(w, nil, err)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
