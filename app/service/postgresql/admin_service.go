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

//////////////////////////////////////////////////
// GET ALL USERS (ADMIN ONLY)
//////////////////////////////////////////////////

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

//////////////////////////////////////////////////
// GET USER BY ID
//////////////////////////////////////////////////

func (s *AdminService) GetUserByID(c *fiber.Ctx) error {
    id := c.Params("id")
    userID := c.Locals("user_id").(uuid.UUID)
    role := c.Locals("role_name").(string)

    paramID, err := uuid.Parse(id)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid user id"})
    }

    // User biasa hanya boleh lihat dirinya sendiri
    if role != "admin" && paramID != userID {
        return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
    }

    user, err := s.adminRepo.GetUserByID(paramID)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "user not found"})
    }

    return c.JSON(user)
}

//////////////////////////////////////////////////
// CREATE USER (ADMIN ONLY)
//////////////////////////////////////////////////

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

//////////////////////////////////////////////////
// UPDATE USER
//////////////////////////////////////////////////

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

//////////////////////////////////////////////////
// DELETE USER
//////////////////////////////////////////////////

func (s *AdminService) DeleteUser(c *fiber.Ctx) error {
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

    if err := s.adminRepo.DeleteUser(targetID); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "user deleted"})
}

//////////////////////////////////////////////////
// ASSIGN ROLE (ADMIN ONLY)
//////////////////////////////////////////////////

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
