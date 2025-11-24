package service

import (
    "time"
    "golang.org/x/crypto/bcrypt"
    models "student-performance-report/app/models/postgresql"
    repo "student-performance-report/app/repository/postgresql"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

type AdminService struct {
    adminRepo repo.AdminRepository
    userRepo  repo.UserRepository
}

func NewAdminService(adminRepo repo.AdminRepository, userRepo repo.UserRepository) *AdminService {
    return &AdminService{adminRepo: adminRepo, userRepo: userRepo}
}

// GetAllUsers godoc
// @Summary Get All Users
// @Description Get list of all users (Admin only)
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.User
// @Failure 403,500 {object} map[string]interface{}
// @Router /users [get]
func (s *AdminService) GetAllUsers(c *fiber.Ctx) error {
    role := c.Locals("role_name").(string)

    if role != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "admin only"})
    }

    users, err := s.adminRepo.GetAllUsers()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(users)
}

// GetUserByID godoc
// @Summary Get User by ID
// @Description Get user details by ID
// @Tags Users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User UUID"
// @Success 200 {object} models.User
// @Failure 400,403,404 {object} map[string]interface{}
// @Router /users/{id} [get]
func (s *AdminService) GetUserByID(c *fiber.Ctx) error {
    id := c.Params("id")
    userID := c.Locals("user_id").(uuid.UUID)
    role := c.Locals("role_name").(string)

    paramID, err := uuid.Parse(id)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
    }

    if role != "admin" && paramID != userID {
        return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
    }

    user, err := s.adminRepo.GetUserByID(paramID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "user not found"})
    }

    return c.JSON(user)
}

// CreateUser godoc
// @Summary Create New User
// @Description Create a new user (Admin only)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body models.User true "User Data"
// @Success 200 {object} models.User
// @Failure 400,403,500 {object} map[string]interface{}
// @Router /users [post]
func (s *AdminService) CreateUser(c *fiber.Ctx) error {
    role := c.Locals("role_name").(string)

    if role != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "admin only"})
    }

    var req models.User
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
    }

    req.ID = uuid.New()
    req.CreatedAt = time.Now()
    req.UpdatedAt = time.Now()

    hashed, _ := bcrypt.GenerateFromPassword([]byte(req.PasswordHash), bcrypt.DefaultCost)
    req.PasswordHash = string(hashed)

    if err := s.adminRepo.CreateUser(&req); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(req)
}

// UpdateUser godoc
// @Summary Update User
// @Description Update user data
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User UUID"
// @Param request body models.User true "User Data"
// @Success 200 {object} models.User
// @Failure 400,403,500 {object} map[string]interface{}
// @Router /users/{id} [put]
func (s *AdminService) UpdateUser(c *fiber.Ctx) error {
    paramID := c.Params("id")
    userID := c.Locals("user_id").(uuid.UUID)
    role := c.Locals("role_name").(string)

    targetID, err := uuid.Parse(paramID)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
    }

    if role != "admin" && targetID != userID {
        return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
    }

    var req models.User
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
    }

    req.ID = targetID

    if err := s.adminRepo.UpdateUser(&req); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(req)
}

// DeleteUser godoc
// @Summary Delete User
// @Description Soft delete user
// @Tags Users
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} map[string]string
// @Failure 400,403,500 {object} map[string]interface{}
// @Router /users/{id} [delete]
func (s *AdminService) DeleteUser(c *fiber.Ctx) error {
	paramID := c.Params("id")
	targetID, err := uuid.Parse(paramID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
	}

	userID := c.Locals("user_id").(uuid.UUID)
	role := c.Locals("role_name").(string)

	if role != "admin" && userID != targetID {
		return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
	}

	if err := s.adminRepo.DeleteUser(targetID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user deactivated (soft deleted)"})
}

// AssignRole godoc
// @Summary Assign Role to User
// @Description Change user role (Admin only)
// @Tags Users
// @Security BearerAuth
// @Accept json
// @Param id path string true "User UUID"
// @Param request body object{roleId=string} true "Role ID"
// @Success 200 {object} map[string]string
// @Failure 400,403,500 {object} map[string]interface{}
// @Router /users/{id}/role [put]
func (s *AdminService) AssignRole(c *fiber.Ctx) error {
    role := c.Locals("role_name").(string)
    if role != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "admin only"})
    }

    var req struct {
        RoleID string `json:"roleId"`
    }
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
    }

    userID, _ := uuid.Parse(c.Params("id"))
    roleID, _ := uuid.Parse(req.RoleID)

    if err := s.adminRepo.AssignRole(userID, roleID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "role assigned"})
}
