package middleware

import (
	"strings"
	"student-performance-report/utils"
	"github.com/gofiber/fiber/v2"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(401).JSON(fiber.Map{"error": "missing token"})
		}

		parts := strings.Split(auth, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token format"})
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role_id", claims.RoleID)
		c.Locals("role_name", claims.RoleName) // penting!
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}

func RoleAllowed(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role_name")
		if role == nil {
			return c.Status(403).JSON(fiber.Map{"error": "role missing"})
		}

		userRole := role.(string)

		for _, r := range allowedRoles {
			if userRole == r {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "forbidden: role not allowed"})
	}
}


func PermissionRequired(needed string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Locals("permissions")
		if raw == nil {
			return c.Status(403).JSON(fiber.Map{"error": "no permissions"})
		}

		perms := raw.([]string)

		for _, p := range perms {
			if p == needed {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "permission denied"})
	}
}

