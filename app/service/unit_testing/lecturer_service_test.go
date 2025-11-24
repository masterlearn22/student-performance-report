package service_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	models "student-performance-report/app/models/postgresql"
	"student-performance-report/app/repository/mocks"
	"student-performance-report/app/service/postgresql"
)

// --- SETUP HELPERS ---

func setupLecturerServiceTest() (*service.LecturerService, *mocks.MockLecturerRepo) {
	mockLecturerRepo := new(mocks.MockLecturerRepo)
	svc := service.NewLecturerService(mockLecturerRepo)
	return svc, mockLecturerRepo
}

func setupSimpleApp() *fiber.App {
	return fiber.New()
}

// --- TEST CASES ---

func TestGetAllLecturers(t *testing.T) {
	t.Run("Success: Get All Lecturers", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		mockData := []models.Lecturer{
			{ID: uuid.New(), LecturerID: "D001", Department: "CS"},
			{ID: uuid.New(), LecturerID: "D002", Department: "IT"},
		}

		mockRepo.On("GetAllLecturers").Return(mockData, nil)

		app.Get("/lecturers", svc.GetAllLecturers)

		req := httptest.NewRequest("GET", "/lecturers", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error: Database Error", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		mockRepo.On("GetAllLecturers").Return(nil, errors.New("db error"))

		app.Get("/lecturers", svc.GetAllLecturers)

		req := httptest.NewRequest("GET", "/lecturers", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})
}

func TestGetLecturerByID(t *testing.T) {
	t.Run("Success: Get Lecturer Found", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		targetID := uuid.New()
		mockLecturer := &models.Lecturer{ID: targetID, LecturerID: "D123"}

		mockRepo.On("GetLecturerByID", targetID).Return(mockLecturer, nil)

		app.Get("/lecturers/:id", svc.GetLecturerByID)

		req := httptest.NewRequest("GET", "/lecturers/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error: Lecturer Not Found", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		targetID := uuid.New()

		mockRepo.On("GetLecturerByID", targetID).Return(nil, errors.New("lecturer not found"))

		app.Get("/lecturers/:id", svc.GetLecturerByID)

		req := httptest.NewRequest("GET", "/lecturers/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 404, resp.StatusCode)
	})
}

func TestGetAdvisees(t *testing.T) {
	t.Run("Success: Get Advisees", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		lecturerID := uuid.New()
		mockStudents := []models.Student{
			{ID: uuid.New(), StudentID: "S1"},
			{ID: uuid.New(), StudentID: "S2"},
		}

		mockRepo.On("GetAdvisees", lecturerID).Return(mockStudents, nil)

		app.Get("/lecturers/:id/advisees", svc.GetAdvisees)

		req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error: Database Error", func(t *testing.T) {
		svc, mockRepo := setupLecturerServiceTest()
		app := setupSimpleApp()

		lecturerID := uuid.New()

		mockRepo.On("GetAdvisees", lecturerID).Return(nil, errors.New("db fail"))

		app.Get("/lecturers/:id/advisees", svc.GetAdvisees)

		req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})
}