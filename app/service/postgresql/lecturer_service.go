package service

import (
	repo "student-performance-report/app/repository/postgresql"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LecturerService struct {
	lecturerRepo repo.LecturerRepository
}

func NewLecturerService(r repo.LecturerRepository) *LecturerService {
	return &LecturerService{lecturerRepo: r}
}

// GetAllLecturers godoc
// @Summary Get All Lecturers
// @Description Get list of all lecturers
// @Tags Students & Lecturers
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Lecturer
// @Router /lecturers [get]
func (s *LecturerService) GetAllLecturers(c *fiber.Ctx) error {
	data, err := s.lecturerRepo.GetAllLecturers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(data)
}

func (s *LecturerService) GetLecturerByID(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))

	lecturer, err := s.lecturerRepo.GetLecturerByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "lecturer not found"})
	}

	return c.JSON(lecturer)
}

// GetAdvisees godoc
// @Summary Get Lecturer Advisees
// @Description Get list of students advised by this lecturer
// @Tags Students & Lecturers
// @Security BearerAuth
// @Param id path string true "Lecturer UUID"
// @Produce json
// @Success 200 {array} models.Student
// @Router /lecturers/{id}/advisees [get]
func (s *LecturerService) GetAdvisees(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))

	students, err := s.lecturerRepo.GetAdvisees(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(students)
}

