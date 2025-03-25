package db

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

func Register(ctx context.Context, db *pgxpool.Pool, email, password, displayName string) (*User, error) {

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// User object
	user := User{
		ID:          uuid.New(),
		Email:       email,
		Password:    string(hashedPassword),
		DisplayName: displayName,
	}
	// Insert into database & return user ID
	err = db.QueryRow(ctx, `
		INSERT INTO users (id, email, password, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, user.ID, user.Email, user.Password, user.DisplayName).Scan(&user.ID)

	if err != nil {
		return nil, err
	}

	return &user, nil

}

// GenerateToken generates a JWT token
func GenerateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(30 * 24 * time.Hour).Unix(), // Token lasts for 30 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func Login(ctx context.Context, db *pgxpool.Pool, email, password string) (string, error) {
	var user User

	// Fetch user from db
	err := db.QueryRow(ctx, "SELECT id, email, password FROM users WHERE email=$1", email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// Generate token
	token, err := GenerateToken(user.ID)
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return token, nil
}
