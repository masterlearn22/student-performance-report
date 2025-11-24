package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	// Import Models
	models "student-performance-report/app/models/postgresql"
	// Import Interface Repo
	repoPg "student-performance-report/app/repository/postgresql"
)

type MockLecturerRepo struct {
	mock.Mock
}

// Compile-time check: Pastikan MockLecturerRepo mengimplementasikan interface LecturerRepository
var _ repoPg.LecturerRepository = (*MockLecturerRepo)(nil)

// --- Method yang sebelumnya sudah ada ---

func (m *MockLecturerRepo) GetLecturerByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockLecturerRepo) GetAdvisees(lecturerID uuid.UUID) ([]models.Student, error) {
	args := m.Called(lecturerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Student), args.Error(1)
}

// --- Method BARU yang WAJIB DITAMBAHKAN (Penyebab Error) ---

func (m *MockLecturerRepo) GetLecturerByID(id uuid.UUID) (*models.Lecturer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lecturer), args.Error(1)
}

func (m *MockLecturerRepo) GetAllLecturers() ([]models.Lecturer, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Lecturer), args.Error(1)
}