
package models

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

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
