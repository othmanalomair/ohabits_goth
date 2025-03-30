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

// AuthMiddleware verifies JWT tokens. It first checks for a token in a cookie named "token".
// If not found, it checks the Authorization header. If no valid token is found, it redirects to /login.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Try to get token from the cookie.
		cookie, err := r.Cookie("token")
		if err == nil && cookie.Value != "" {
			tokenString = cookie.Value
		} else {
			// Fallback to the Authorization header.
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
		}

		// Parse and validate the token.
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Extract the user_id from the claims.
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store the user ID in the context so that handlers can access it.
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
