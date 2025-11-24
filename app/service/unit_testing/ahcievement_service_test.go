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
	"github.com/stretchr/testify/mock"

	modelMongo "student-performance-report/app/models/mongodb"
	modelPg "student-performance-report/app/models/postgresql"
	"student-performance-report/app/repository/mocks"
	"student-performance-report/app/service/mongodb"
)

// --- SETUP HELPERS ---

func setupAchievementServiceTest() (*service.AchievementService, *mocks.MockAchievementMongoRepo, *mocks.MockAchievementPgRepo, *mocks.MockLecturerRepo) {
	mockMongo := new(mocks.MockAchievementMongoRepo)
	mockPg := new(mocks.MockAchievementPgRepo)
	mockLecturer := new(mocks.MockLecturerRepo)

	svc := service.NewAchievementService(mockMongo, mockPg, mockLecturer)

	return svc, mockMongo, mockPg, mockLecturer
}

func setupAchievementApp(roleName string, userID uuid.UUID) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("role_name", roleName)
		c.Locals("user_id", userID)
		return c.Next()
	})
	return app
}

// --- TEST CASES ---

func TestCreateAchievement(t *testing.T) {
	t.Run("Success: Create Draft Achievement", func(t *testing.T) {
		svc, mockMongo, mockPg, _ := setupAchievementServiceTest()
		userID := uuid.New()
		app := setupAchievementApp("mahasiswa", userID)

		studentID := uuid.New()
		mongoID := "mongo_obj_id_123"
		newRefID := uuid.New()

		reqBody := modelMongo.Achievement{
			Title:       "Lomba Coding",
			Description: "Juara 1",
		}

		// 1. Mock GetStudentByUserID (PG)
		mockPg.On("GetStudentByUserID", mock.Anything, userID).Return(studentID, nil)

		// 2. Mock InsertOne (Mongo) - Menyimpan detail
		mockMongo.On("InsertOne", mock.Anything, mock.MatchedBy(func(a modelMongo.Achievement) bool {
			return a.Title == "Lomba Coding" && a.StudentID == studentID.String()
		})).Return(mongoID, nil)

		// 3. Mock Create (PG) - Menyimpan referensi
		mockPg.On("Create", mock.Anything, mock.MatchedBy(func(r modelPg.AchievementReference) bool {
			return r.StudentID == studentID && r.MongoAchievementID == mongoID && r.Status == "draft"
		})).Return(newRefID, nil)

		app.Post("/achievements", svc.CreateAchievement)

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/achievements", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, 201, resp.StatusCode)
		mockPg.AssertExpectations(t)
		mockMongo.AssertExpectations(t)
	})

	t.Run("Error: Student Profile Not Found", func(t *testing.T) {
		svc, _, mockPg, _ := setupAchievementServiceTest()
		userID := uuid.New()
		app := setupAchievementApp("mahasiswa", userID)

		mockPg.On("GetStudentByUserID", mock.Anything, userID).Return(uuid.Nil, errors.New("not found"))

		app.Post("/achievements", svc.CreateAchievement)

		req := httptest.NewRequest("POST", "/achievements", nil)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req)

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestSubmitAchievement(t *testing.T) {
	t.Run("Success: Submit Draft", func(t *testing.T) {
		svc, _, mockPg, _ := setupAchievementServiceTest()
		userID := uuid.New()
		app := setupAchievementApp("mahasiswa", userID)

		achievementID := uuid.New()
		studentID := uuid.New()

		// Data Mock Reference
		ref := modelPg.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "draft",
		}

		// 1. Get Student ID check
		mockPg.On("GetStudentByUserID", mock.Anything, userID).Return(studentID, nil)

		// 2. Get Reference Check
		mockPg.On("GetReferenceByID", mock.Anything, achievementID).Return(ref, nil)

		// 3. Submit Action
		mockPg.On("SubmitReference", mock.Anything, achievementID).Return(nil)

		app.Post("/achievements/:id/submit", svc.SubmitAchievement)

		req := httptest.NewRequest("POST", "/achievements/"+achievementID.String()+"/submit", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockPg.AssertExpectations(t)
	})

	t.Run("Error: Cannot Submit Non-Draft", func(t *testing.T) {
		svc, _, mockPg, _ := setupAchievementServiceTest()
		userID := uuid.New()
		app := setupAchievementApp("mahasiswa", userID)

		achievementID := uuid.New()
		studentID := uuid.New()

		ref := modelPg.AchievementReference{
			ID:        achievementID,
			StudentID: studentID,
			Status:    "verified", // Status bukan draft
		}

		mockPg.On("GetStudentByUserID", mock.Anything, userID).Return(studentID, nil)
		mockPg.On("GetReferenceByID", mock.Anything, achievementID).Return(ref, nil)

		app.Post("/achievements/:id/submit", svc.SubmitAchievement)

		req := httptest.NewRequest("POST", "/achievements/"+achievementID.String()+"/submit", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 400, resp.StatusCode) // Bad Request
	})
}

func TestVerifyAchievement(t *testing.T) {
	t.Run("Success: Lecturer Verifies Achievement", func(t *testing.T) {
		svc, _, mockPg, mockLecturer := setupAchievementServiceTest()
		lecturerUserID := uuid.New()
		achievementID := uuid.New()
		
		app := setupAchievementApp("dosen_wali", lecturerUserID)

		ref := modelPg.AchievementReference{
			ID:     achievementID,
			Status: "submitted",
		}

		// 1. Check is Lecturer
		mockLecturer.On("GetLecturerByUserID", mock.Anything, lecturerUserID).Return(uuid.New(), nil)

		// 2. Get Reference
		mockPg.On("GetReferenceByID", mock.Anything, achievementID).Return(ref, nil)

		// 3. Update Status
		mockPg.On("UpdateStatus", mock.Anything, achievementID, "verified", &lecturerUserID, "").Return(nil)

		app.Post("/achievements/:id/verify", svc.VerifyAchievement)

		req := httptest.NewRequest("POST", "/achievements/"+achievementID.String()+"/verify", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
	})
}