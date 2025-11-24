package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt" 
	models "student-performance-report/app/models/postgresql"
	"student-performance-report/app/repository/mocks"
	"student-performance-report/app/service/postgresql"
)

// --- SETUP HELPERS ---

func setupAuthServiceTest() (*service.AuthService, *mocks.MockUserRepo) {
	mockUserRepo := new(mocks.MockUserRepo)
	svc := service.NewAuthService(mockUserRepo)
	return svc, mockUserRepo
}

func setupAuthApp() *fiber.App {
	return fiber.New()
}

// --- TEST CASES ---

func TestLogin(t *testing.T) {
	t.Run("Success: Login with valid credentials", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := setupAuthApp()

		// 1. Siapkan Password Hash yang VALID
		passwordRaw := "password123"
		hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(passwordRaw), bcrypt.DefaultCost)
		passwordHash := string(hashedBytes)

		roleID := uuid.New()
		mockUser := &models.User{
			ID:           uuid.New(),
			Username:     "admin",
			PasswordHash: passwordHash, 
			FullName:     "Admin User",
			RoleID:       roleID,
			IsActive:     true,
		}
		roleName := "admin"
		permissions := []string{"read_users", "write_users"}

		// 2. Mock Expectations
		mockRepo.On("GetByUsername", "admin").Return(mockUser, roleName, nil)
		mockRepo.On("GetPermissionsByRoleID", roleID).Return(permissions, nil)

		// 3. Execute
		app.Post("/login", svc.Login)

		payload := map[string]string{
			"username": "admin",
			"password": passwordRaw, 
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		// 4. Assertions
		assert.Equal(t, 200, resp.StatusCode)
		
		// Cek apakah token ada di response
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.NotEmpty(t, response["token"])
		assert.NotEmpty(t, response["refreshToken"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error: Invalid Password", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := setupAuthApp()

		// Hash password "rahasia"
		hashedBytes, _ := bcrypt.GenerateFromPassword([]byte("rahasia"), bcrypt.DefaultCost)
		
		mockUser := &models.User{
			Username:     "user1",
			PasswordHash: string(hashedBytes),
			IsActive:     true,
		}

		// Mock return user sukses
		mockRepo.On("GetByUsername", "user1").Return(mockUser, "student", nil)

		app.Post("/login", svc.Login)

		// Tapi input password SALAH
		payload := map[string]string{
			"username": "user1",
			"password": "password_salah", 
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Error: User Not Found", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := setupAuthApp()

		mockRepo.On("GetByUsername", "unknown").Return(nil, "", errors.New("user not found"))

		app.Post("/login", svc.Login)

		payload := map[string]string{"username": "unknown", "password": "123"}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Error: Inactive Account", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := setupAuthApp()

		passwordRaw := "pass123"
		hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(passwordRaw), bcrypt.DefaultCost)

		mockUser := &models.User{
			Username:     "inactive_user",
			PasswordHash: string(hashedBytes),
			IsActive:     false, 
		}

		mockRepo.On("GetByUsername", "inactive_user").Return(mockUser, "student", nil)

		app.Post("/login", svc.Login)

		payload := map[string]string{"username": "inactive_user", "password": passwordRaw}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, 403, resp.StatusCode) 
	})
}

func TestProfile(t *testing.T) {
	t.Run("Success: Get Profile", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := fiber.New()
		
		userID := uuid.New()
		roleID := uuid.New()

		// Middleware simulasi Auth (inject user_id ke locals)
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_id", userID)
			return c.Next()
		})

		mockUser := &models.User{
			ID:       userID,
			Username: "myprofile",
			FullName: "My Name",
			RoleID:   roleID,
		}
		permissions := []string{"read_profile"}

		// Expectations
		mockRepo.On("GetByID", userID).Return(mockUser, nil)
		mockRepo.On("GetPermissionsByRoleID", roleID).Return(permissions, nil)
		mockRepo.On("GetByUsername", "myprofile").Return(mockUser, "student", nil) 

		app.Get("/profile", svc.Profile)

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		
		var respBody models.UserResp
		json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, "myprofile", respBody.Username)
	})

	t.Run("Error: User Not Found (ID from Token invalid in DB)", func(t *testing.T) {
		svc, mockRepo := setupAuthServiceTest()
		app := fiber.New()
		userID := uuid.New()

		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_id", userID)
			return c.Next()
		})

		mockRepo.On("GetByID", userID).Return(nil, errors.New("not found"))

		app.Get("/profile", svc.Profile)

		req := httptest.NewRequest("GET", "/profile", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestLogout(t *testing.T) {
	t.Run("Success: Logout", func(t *testing.T) {
		svc, _ := setupAuthServiceTest()
		app := setupAuthApp()

		app.Post("/logout", svc.Logout)

		req := httptest.NewRequest("POST", "/logout", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
	})
}