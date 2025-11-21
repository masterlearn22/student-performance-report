package service

import (


	models "student-performance-report/app/models/postgresql"
	repo "student-performance-report/app/repository/postgresql"
	"student-performance-report/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type authService struct {
	userRepo repo.UserRepository
}

func NewAuthService(userRepo repo.UserRepository) *authService {
	return &authService{userRepo: userRepo}
}


func (s *authService) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// Ambil JSON request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}

	// Cari user berdasarkan username
	user, roleName, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	// Cek password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	// Cek apakah user aktif
	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "account is inactive"})
	}

	// Ambil permissions user
	permissions, err := s.userRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate token
	tokenString, err := utils.GenerateToken(user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate refresh token
	refresh, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Response
	return c.JSON(models.LoginResponse{
		Token:        tokenString,
		RefreshToken: refresh,
		User: models.UserResp{
			ID:          user.ID,
			Username:    user.Username,
			FullName:    user.FullName,
			Role:        roleName,
			Permissions: permissions,
		},
	})
}

func (s *authService) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	permissions, _ := s.userRepo.GetPermissionsByRoleID(user.RoleID)
	_, roleName, _ := s.userRepo.GetByUsername(user.Username)

	newToken, err := utils.GenerateToken(user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"token": newToken})
}

func (s *authService) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logout successful",
	})
}

func (s *authService) Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	id, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	permissions, _ := s.userRepo.GetPermissionsByRoleID(user.RoleID)
	_, roleName, _ := s.userRepo.GetByUsername(user.Username)

	return c.JSON(models.UserResp{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Role:        roleName,
		Permissions: permissions,
	})
}


