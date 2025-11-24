package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	// Import Models
	modelMongo "student-performance-report/app/models/mongodb"
	models "student-performance-report/app/models/postgresql"

	// Import Interfaces
	repoMongo "student-performance-report/app/repository/mongodb"
	repoPg "student-performance-report/app/repository/postgresql"
)

// =========================================================
// MOCK STUDENT REPOSITORY (PostgreSQL)
// =========================================================

type MockStudentRepo struct {
	mock.Mock
}

// Compile-time check
var _ repoPg.StudentRepository = (*MockStudentRepo)(nil)

func (m *MockStudentRepo) GetAllStudents(ctx context.Context) ([]models.Student, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Student), args.Error(1)
}

func (m *MockStudentRepo) GetStudentByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Student), args.Error(1)
}

func (m *MockStudentRepo) UpdateAdvisor(ctx context.Context, studentID, lecturerID uuid.UUID) error {
	args := m.Called(ctx, studentID, lecturerID)
	return args.Error(0)
}

func (m *MockStudentRepo) GetStudentsByIDs(ctx context.Context, ids []string) ([]models.StudentWithUser, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StudentWithUser), args.Error(1)
}

// =========================================================
// MOCK ACHIEVEMENT REPOSITORY (MongoDB)
// =========================================================
// Perhatikan: Nama struct disesuaikan dengan yang dipanggil di test file Anda
// yaitu "MockAchievementRepo" (bukan MockAchievementMongoRepo)

type MockAchievementRepo struct {
	mock.Mock
}

// Compile-time check: Pastikan struct ini memenuhi Interface AchievementRepository
var _ repoMongo.AchievementRepository = (*MockAchievementRepo)(nil)

// 1. Method yang dipakai di StudentService
func (m *MockAchievementRepo) GetStudentAchievements(studentId uuid.UUID) ([]modelMongo.Achievement, error) {
	args := m.Called(studentId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]modelMongo.Achievement), args.Error(1)
}

// 2. Method lain yang WAJIB ada (Implementasi Interface) walau tidak dipakai di test ini
func (m *MockAchievementRepo) InsertOne(ctx context.Context, achievement modelMongo.Achievement) (string, error) {
	args := m.Called(ctx, achievement)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepo) FindAllDetails(ctx context.Context, mongoIDs []string) ([]modelMongo.Achievement, error) {
	args := m.Called(ctx, mongoIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]modelMongo.Achievement), args.Error(1)
}

func (m *MockAchievementRepo) FindOne(ctx context.Context, mongoID string) (*modelMongo.Achievement, error) {
	args := m.Called(ctx, mongoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.Achievement), args.Error(1)
}

func (m *MockAchievementRepo) DeleteAchievement(ctx context.Context, mongoID string) error {
	args := m.Called(ctx, mongoID)
	return args.Error(0)
}

func (m *MockAchievementRepo) UpdateOne(ctx context.Context, mongoID string, data modelMongo.Achievement) error {
	args := m.Called(ctx, mongoID, data)
	return args.Error(0)
}

// INI METHOD PENYEBAB ERROR (Sudah ditambahkan)
func (m *MockAchievementRepo) AddAttachment(ctx context.Context, mongoID string, attachment modelMongo.Attachment) error {
	args := m.Called(ctx, mongoID, attachment)
	return args.Error(0)
}

func (m *MockAchievementRepo) GetGlobalStats(ctx context.Context) (*modelMongo.GlobalStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.GlobalStatistics), args.Error(1)
}

func (m *MockAchievementRepo) GetStudentStats(ctx context.Context, studentID string) (*modelMongo.StudentStatistics, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.StudentStatistics), args.Error(1)
}