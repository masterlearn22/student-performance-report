package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims adalah struct untuk payload token
// Kita embed jwt.RegisteredClaims untuk field standar (Exp, Iss, dll)
type JWTClaims struct {
	UserID      uuid.UUID `json:"userId"`
	RoleID      uuid.UUID `json:"roleId"`
	RoleName    string    `json:"roleName"`
	Permissions []string  `json:"permissions,omitempty"` 
	
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

// Ini Struct tambahan untuk Response Login (DTO)
// Sesuai contoh response di SRS Appendix Poin 2 [cite: 312-325]
type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	User         UserResp `json:"user"`
}

type UserResp struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	FullName    string    `json:"fullName"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
}

