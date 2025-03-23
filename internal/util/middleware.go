package util

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
}

var jwtSecret []byte

func init() {
	LoadEnv()
	jwtSecret = []byte(getJWTSecret()) // Initialize once
}

func getJWTSecret() string {
	LoadEnv()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}
	return secret
}

// AuthMiddleware verifies JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user ID as UUID
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Convert user ID to UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Error(w, "Invalid UUID", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
