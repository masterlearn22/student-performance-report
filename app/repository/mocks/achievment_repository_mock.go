package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	// Import Models
	modelMongo "student-performance-report/app/models/mongodb"
	modelPg "student-performance-report/app/models/postgresql"

	// Import Interfaces
	repoMongo "student-performance-report/app/repository/mongodb"
	repoPg "student-performance-report/app/repository/postgresql"
)

// =========================================================
// MOCK ACHIEVEMENT REPOSITORY (MongoDB)
// =========================================================

// PERHATIKAN: Sesuaikan nama struct ini dengan yang dipanggil di Test File Anda.
// Jika di test file Anda memanggil "MockAchievementRepo", ubah nama struct ini jadi "MockAchievementRepo"
type MockAchievementMongoRepo struct {
	mock.Mock
}

// Compile-time check implementation
var _ repoMongo.AchievementRepository = (*MockAchievementMongoRepo)(nil)

func (m *MockAchievementMongoRepo) GetStudentAchievements(studentId uuid.UUID) ([]modelMongo.Achievement, error) {
	args := m.Called(studentId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]modelMongo.Achievement), args.Error(1)
}

func (m *MockAchievementMongoRepo) InsertOne(ctx context.Context, achievement modelMongo.Achievement) (string, error) {
	args := m.Called(ctx, achievement)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementMongoRepo) FindAllDetails(ctx context.Context, mongoIDs []string) ([]modelMongo.Achievement, error) {
	args := m.Called(ctx, mongoIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]modelMongo.Achievement), args.Error(1)
}

func (m *MockAchievementMongoRepo) FindOne(ctx context.Context, mongoID string) (*modelMongo.Achievement, error) {
	args := m.Called(ctx, mongoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.Achievement), args.Error(1)
}

func (m *MockAchievementMongoRepo) DeleteAchievement(ctx context.Context, mongoID string) error {
	args := m.Called(ctx, mongoID)
	return args.Error(0)
}

func (m *MockAchievementMongoRepo) UpdateOne(ctx context.Context, mongoID string, data modelMongo.Achievement) error {
	args := m.Called(ctx, mongoID, data)
	return args.Error(0)
}

func (m *MockAchievementMongoRepo) AddAttachment(ctx context.Context, mongoID string, attachment modelMongo.Attachment) error {
	args := m.Called(ctx, mongoID, attachment)
	return args.Error(0)
}

func (m *MockAchievementMongoRepo) GetGlobalStats(ctx context.Context) (*modelMongo.GlobalStatistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.GlobalStatistics), args.Error(1)
}

func (m *MockAchievementMongoRepo) GetStudentStats(ctx context.Context, studentID string) (*modelMongo.StudentStatistics, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modelMongo.StudentStatistics), args.Error(1)
}

// =========================================================
// MOCK ACHIEVEMENT REPOSITORY (PostgreSQL)
// =========================================================

type MockAchievementPgRepo struct {
	mock.Mock
}

// Compile-time check implementation
var _ repoPg.AchievementRepoPostgres = (*MockAchievementPgRepo)(nil)

func (m *MockAchievementPgRepo) Create(ctx context.Context, ref modelPg.AchievementReference) (uuid.UUID, error) {
	args := m.Called(ctx, ref)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAchievementPgRepo) GetStudentByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAchievementPgRepo) GetAllReferences(ctx context.Context, filter map[string]interface{}, limit, offset int, sort string) ([]modelPg.AchievementReference, int64, error) {
	args := m.Called(ctx, filter, limit, offset, sort)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]modelPg.AchievementReference), args.Get(1).(int64), args.Error(2)
}

func (m *MockAchievementPgRepo) GetReferenceByID(ctx context.Context, id uuid.UUID) (modelPg.AchievementReference, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(modelPg.AchievementReference), args.Error(1)
}

func (m *MockAchievementPgRepo) DeleteReference(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAchievementPgRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, note string) error {
	args := m.Called(ctx, id, status, verifiedBy, note)
	return args.Error(0)
}

func (m *MockAchievementPgRepo) SubmitReference(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

