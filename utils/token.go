package utils

import (
	"os"
	"time"
	"errors"
	// "fmt"

	"student-performance-report/app/models/postgresql"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken membuat string token JWT lengkap dengan claims
func GenerateToken(user *models.User, roleName string, permissions []string) (string, error) {
	// Ambil secret dari .env
	secret := os.Getenv("JWT_SECRET")
	
	// Set waktu expired (misal 24 jam)
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &models.JWTClaims{
		UserID:      user.ID,
		RoleID:      user.RoleID,
		RoleName:    roleName,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "student-performance-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string) (*models.JWTClaims, error) {
	secret := os.Getenv("JWT_SECRET")

	token, err := jwt.ParseWithClaims(
		tokenString,
		&models.JWTClaims{}, // pastikan pakai struct yang mengandung RoleName
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// DEBUG
	// fmt.Println("RoleName:", claims.RoleName)

	return claims, nil
}



func GenerateRefreshToken(user *models.User) (string, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET") // fallback
	}

	// Refresh token: masa berlaku 7 hari
	expiration := time.Now().Add(7 * 24 * time.Hour)

	claims := &models.RefreshClaims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			Issuer:    "student-performance-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateRefreshToken(t string) (*models.RefreshClaims, error) {
	secret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}

	token, err := jwt.ParseWithClaims(
		t,
		&models.RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return claims, nil
}

