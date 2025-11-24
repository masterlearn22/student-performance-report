package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	// Import Models & Service
	models "student-performance-report/app/models/postgresql"
	"student-performance-report/app/service/postgresql"
	
	// Import Mocks
	"student-performance-report/app/repository/mocks"
	modelMongo "student-performance-report/app/models/mongodb"
)

// --- SETUP HELPERS ---

func setupStudentServiceTest() (*service.StudentService, *mocks.MockStudentRepo, *mocks.MockAchievementRepo) {
	mockStudentRepo := new(mocks.MockStudentRepo)
	mockAchievementRepo := new(mocks.MockAchievementRepo)

	// Inject kedua mock repo ke service
	svc := service.NewStudentService(mockStudentRepo, mockAchievementRepo)

	return svc, mockStudentRepo, mockAchievementRepo
}

func setupStudentApp() *fiber.App {
	app := fiber.New()
	return app
}

// --- TEST CASES ---

func TestGetAllStudents(t *testing.T) {
	t.Run("Success: Get all students", func(t *testing.T) {
		svc, mockStudentRepo, _ := setupStudentServiceTest()
		app := setupStudentApp()

		// Dummy Data
		mockData := []models.Student{
			{ID: uuid.New(), StudentID: "12345"},
			{ID: uuid.New(), StudentID: "67890"},
		}

		// Expectation: Context apa saja (mock.Anything), return mockData
		mockStudentRepo.On("GetAllStudents", mock.Anything).Return(mockData, nil)

		app.Get("/students", svc.GetAllStudents)

		req := httptest.NewRequest("GET", "/students", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockStudentRepo.AssertExpectations(t)
	})

	t.Run("Error: Database Failure", func(t *testing.T) {
		svc, mockStudentRepo, _ := setupStudentServiceTest()
		app := setupStudentApp()

		mockStudentRepo.On("GetAllStudents", mock.Anything).Return(nil, errors.New("db error"))

		app.Get("/students", svc.GetAllStudents)

		req := httptest.NewRequest("GET", "/students", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})
}

func TestGetStudentByID(t *testing.T) {
	t.Run("Success: Get student by ID", func(t *testing.T) {
		svc, mockStudentRepo, _ := setupStudentServiceTest()
		app := setupStudentApp()

		targetID := uuid.New()
		mockStudent := &models.Student{ID: targetID, StudentID: "NIM123"}

		mockStudentRepo.On("GetStudentByID", mock.Anything, targetID).Return(mockStudent, nil)

		app.Get("/students/:id", svc.GetStudentByID)

		req := httptest.NewRequest("GET", "/students/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockStudentRepo.AssertExpectations(t)
	})

	t.Run("Error: Invalid UUID format", func(t *testing.T) {
		svc, _, _ := setupStudentServiceTest()
		app := setupStudentApp()

		app.Get("/students/:id", svc.GetStudentByID)

		req := httptest.NewRequest("GET", "/students/invalid-uuid", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Error: Student Not Found", func(t *testing.T) {
		svc, mockStudentRepo, _ := setupStudentServiceTest()
		app := setupStudentApp()

		targetID := uuid.New()
		mockStudentRepo.On("GetStudentByID", mock.Anything, targetID).Return(nil, errors.New("student not found"))

		app.Get("/students/:id", svc.GetStudentByID)

		req := httptest.NewRequest("GET", "/students/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestGetStudentAchievements(t *testing.T) {
	t.Run("Success: Get achievements", func(t *testing.T) {
		svc, _, mockAchievementRepo := setupStudentServiceTest()
		app := setupStudentApp()

		targetID := uuid.New()
		// Mock data achievement (sesuaikan tipe data ini dengan model mongoDB Anda)
		mockAchievements := []modelMongo.Achievement{
    {
        // Hapus tanda kutip di key, dan gunakan Huruf Besar (sesuai definisi Struct)
        AchievementType: "competition",
        Title:           "Juara 1 Hackathon Nasional Gemastik 2025",
        Description:     "Memenangkan medali emas kategori Pengembangan Perangkat Lunak dalam kompetisi tingkat nasional.",
        Tags:            []string{"coding", "hackathon", "software engineering"},
        Points:          150,
        
        // Field 'details' kemungkinan besar adalah map[string]interface{}
        // Jadi harus diinisialisasi sebagai map:
        Details: modelMongo.AchievementDetails{
            CompetitionName:  "Gemastik 2025",
            CompetitionLevel: "national",
            Rank:             1,
            MedalType:        "gold",
            Location:         "Universitas Indonesia, Jakarta",
            Organizer:        "Puspresnas Kemendikbud",
            Score:            98.5,
            EventDate:        time.Date(2025, 10, 15, 9, 0, 0, 0, time.UTC),
			},
		},
	}

		mockAchievementRepo.On("GetStudentAchievements", targetID).Return(mockAchievements, nil)

		app.Get("/students/:id/achievements", svc.GetStudentAchievements)

		req := httptest.NewRequest("GET", "/students/"+targetID.String()+"/achievements", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockAchievementRepo.AssertExpectations(t)
	})

	t.Run("Error: Repo Failure", func(t *testing.T) {
		svc, _, mockAchievementRepo := setupStudentServiceTest()
		app := setupStudentApp()

		targetID := uuid.New()
		mockAchievementRepo.On("GetStudentAchievements", targetID).Return(nil, errors.New("mongo error"))

		app.Get("/students/:id/achievements", svc.GetStudentAchievements)

		req := httptest.NewRequest("GET", "/students/"+targetID.String()+"/achievements", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})
}

func TestUpdateAdvisor(t *testing.T) {
	t.Run("Success: Update Advisor", func(t *testing.T) {
		svc, mockStudentRepo, _ := setupStudentServiceTest()
		app := setupStudentApp()

		studentID := uuid.New()
		lecturerID := uuid.New()

		// Expectation
		mockStudentRepo.On("UpdateAdvisor", mock.Anything, studentID, lecturerID).Return(nil)

		app.Put("/students/:id/advisor", svc.UpdateAdvisor)

		// Request Body
		bodyPayload := map[string]string{"lecturerId": lecturerID.String()}
		bodyBytes, _ := json.Marshal(bodyPayload)

		req := httptest.NewRequest("PUT", "/students/"+studentID.String()+"/advisor", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockStudentRepo.AssertExpectations(t)
	})

	t.Run("Error: Invalid Lecturer UUID in Body", func(t *testing.T) {
		svc, _, _ := setupStudentServiceTest()
		app := setupStudentApp()
		studentID := uuid.New()

		app.Put("/students/:id/advisor", svc.UpdateAdvisor)

		bodyPayload := map[string]string{"lecturerId": "invalid-uuid"}
		bodyBytes, _ := json.Marshal(bodyPayload)

		req := httptest.NewRequest("PUT", "/students/"+studentID.String()+"/advisor", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		resp, _ := app.Test(req)

		assert.Equal(t, 400, resp.StatusCode)
	})
}