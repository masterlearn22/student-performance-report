package service_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	modelMongo "student-performance-report/app/models/mongodb"
	models "student-performance-report/app/models/postgresql"
	"student-performance-report/app/repository/mocks"
	"student-performance-report/app/service/mongodb"
)

// --- SETUP HELPERS ---

func setupReportServiceTest() (*service.ReportService, *mocks.MockAchievementRepo, *mocks.MockStudentRepo) {
	// Gunakan MockAchievementRepo (MongoDB) dan MockStudentRepo (Postgres)
	mockMongo := new(mocks.MockAchievementRepo)
	mockPg := new(mocks.MockStudentRepo)

	svc := service.NewReportService(mockMongo, mockPg)

	return svc, mockMongo, mockPg
}

func setupReportApp() *fiber.App {
	return fiber.New()
}

// --- TEST CASES ---

func TestGetStatistics(t *testing.T) {
	t.Run("Success: Get Global Stats with Student Details", func(t *testing.T) {
		svc, mockMongo, mockPg := setupReportServiceTest()
		app := setupReportApp()

		// 1. Mock Data dari MongoDB (Global Stats)
		studentUUID := uuid.New()
		mockStats := &modelMongo.GlobalStatistics{
			TypeDistribution:  map[string]int{"competition": 10},
			LevelDistribution: map[string]int{"national": 5},
			PointsDistribution: []modelMongo.TopStudent{
				{StudentID: studentUUID.String(), TotalPoints: 100},
			},
		}

		// 2. Mock Data dari Postgres (Student Details untuk enrichment)
		mockStudentDetails := []models.StudentWithUser{
			{
				ID:           studentUUID,
				FullName:     "Budi Santoso",
				ProgramStudy: "Informatika",
			},
		}

		// Expectation 1: Panggil Mongo GetGlobalStats
		mockMongo.On("GetGlobalStats", mock.Anything).Return(mockStats, nil)

		// Expectation 2: Panggil Postgres GetStudentsByIDs dengan ID dari hasil mongo
		mockPg.On("GetStudentsByIDs", mock.Anything, []string{studentUUID.String()}).Return(mockStudentDetails, nil)

		// Execute
		app.Get("/stats", svc.GetStatistics)
		req := httptest.NewRequest("GET", "/stats", nil)
		resp, _ := app.Test(req)

		// Assertions
		assert.Equal(t, 200, resp.StatusCode)

		// Cek apakah response body mengandung nama mahasiswa (hasil enrichment)
		var responseBody modelMongo.GlobalStatistics
		json.NewDecoder(resp.Body).Decode(&responseBody)
		
		assert.Equal(t, "Budi Santoso", responseBody.PointsDistribution[0].Name)
		assert.Equal(t, "Informatika", responseBody.PointsDistribution[0].ProgramStudy)

		mockMongo.AssertExpectations(t)
		mockPg.AssertExpectations(t)
	})

	t.Run("Error: Mongo DB Failure", func(t *testing.T) {
		svc, mockMongo, _ := setupReportServiceTest()
		app := setupReportApp()

		mockMongo.On("GetGlobalStats", mock.Anything).Return(nil, errors.New("mongo connection failed"))

		app.Get("/stats", svc.GetStatistics)
		req := httptest.NewRequest("GET", "/stats", nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})
}

func TestGetStudentReport(t *testing.T) {
	t.Run("Success: Get Student Report with Profile", func(t *testing.T) {
		svc, mockMongo, mockPg := setupReportServiceTest()
		app := setupReportApp()

		targetID := uuid.New()
		
		// 1. Mock Data Mongo (Stats)
		mockStats := &modelMongo.StudentStatistics{
			TotalPoints:       50,
			TotalAchievements: 5,
			ByType:            map[string]int{"competition": 5},
		}

		// 2. Mock Data Postgres (Profile)
		mockProfile := &models.Student{
			ID:       targetID,
			FullName: "Siti Aminah",
		}

		// Expectation
		// Perhatikan: Service memanggil Mongo dulu menggunakan String ID
		mockMongo.On("GetStudentStats", mock.Anything, targetID.String()).Return(mockStats, nil)
		// Lalu memanggil Postgres menggunakan UUID
		mockPg.On("GetStudentByID", mock.Anything, targetID).Return(mockProfile, nil)

		app.Get("/report/:id", svc.GetStudentReport)
		req := httptest.NewRequest("GET", "/report/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 200, resp.StatusCode)
		
		var body modelMongo.StudentStatistics
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Equal(t, "Siti Aminah", body.StudentName) // Pastikan nama terisi
	})

	t.Run("Error: Mongo DB Failure", func(t *testing.T) {
		svc, mockMongo, _ := setupReportServiceTest()
		app := setupReportApp()
		targetID := uuid.New().String()

		mockMongo.On("GetStudentStats", mock.Anything, targetID).Return(nil, errors.New("db error"))

		app.Get("/report/:id", svc.GetStudentReport)
		req := httptest.NewRequest("GET", "/report/"+targetID, nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 500, resp.StatusCode)
	})

	t.Run("Error: Invalid UUID Format", func(t *testing.T) {
		svc, mockMongo, _ := setupReportServiceTest()
		app := setupReportApp()
		invalidID := "bukan-uuid"

		// Mock Mongo tetap dipanggil karena urutan kode di service: Mongo dulu -> baru Parse UUID
		mockStats := &modelMongo.StudentStatistics{}
		mockMongo.On("GetStudentStats", mock.Anything, invalidID).Return(mockStats, nil)

		app.Get("/report/:id", svc.GetStudentReport)
		req := httptest.NewRequest("GET", "/report/"+invalidID, nil)
		resp, _ := app.Test(req)

		assert.Equal(t, 400, resp.StatusCode)
	})

	t.Run("Success: Student Not Found in Postgres (Return Stats without Name)", func(t *testing.T) {
		svc, mockMongo, mockPg := setupReportServiceTest()
		app := setupReportApp()
		targetID := uuid.New()

		mockStats := &modelMongo.StudentStatistics{TotalPoints: 10}

		mockMongo.On("GetStudentStats", mock.Anything, targetID.String()).Return(mockStats, nil)
		// Mock Postgres return error/not found
		mockPg.On("GetStudentByID", mock.Anything, targetID).Return(nil, errors.New("student not found"))

		app.Get("/report/:id", svc.GetStudentReport)
		req := httptest.NewRequest("GET", "/report/"+targetID.String(), nil)
		resp, _ := app.Test(req)

		// Tetap 200 OK, tapi nama kosong (sesuai logic service)
		assert.Equal(t, 200, resp.StatusCode)
		
		var body modelMongo.StudentStatistics
		json.NewDecoder(resp.Body).Decode(&body)
		assert.Empty(t, body.StudentName)
	})
}