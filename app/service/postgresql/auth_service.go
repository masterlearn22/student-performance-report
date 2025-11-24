package service

import (
	models "student-performance-report/app/models/postgresql"
	repo "student-performance-report/app/repository/postgresql"
	"student-performance-report/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo repo.UserRepository
}

func NewAuthService(userRepo repo.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// Login godoc
// @Summary User Login
// @Description Authenticate user and return token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{username=string,password=string} true "Login Credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400,401,403 {object} map[string]interface{}
// @Router /auth/login [post]
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}

	user, roleName, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{"error": "invalid username or password"})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{"error": "account is inactive"})
	}

	permissions, err := s.userRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	tokenString, err := utils.GenerateToken(user, roleName, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	refresh, err := utils.GenerateRefreshToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

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

// Refresh godoc
// @Summary Refresh Access Token
// @Description Get new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{refreshToken=string} true "Refresh Token"
// @Success 200 {object} map[string]string
// @Failure 400,401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (s *AuthService) Refresh(c *fiber.Ctx) error {
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

// Logout godoc
// @Summary User Logout
// @Description Logout user (Client side should clear token)
// @Tags Authentication
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "logout successful",
	})
}

// Profile godoc
// @Summary Get User Profile
// @Description Get currently logged in user profile
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResp
// @Failure 404 {object} map[string]interface{}
// @Router /auth/profile [get]
func (s *AuthService) Profile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	user, err := s.userRepo.GetByID(userID)
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