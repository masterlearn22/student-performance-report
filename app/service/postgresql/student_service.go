package service

import (
    repo "student-performance-report/app/repository/postgresql"
    mongoRepo "student-performance-report/app/repository/mongodb"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

type StudentService struct {
    studentRepo     repo.StudentRepository
    achievementRepo mongoRepo.AchievementRepository
}

func NewStudentService(r repo.StudentRepository, a mongoRepo.AchievementRepository) *StudentService {
    return &StudentService{studentRepo: r, achievementRepo: a}
}

// GetAllStudents godoc
// @Summary Get All Students
// @Description Get list of all students
// @Tags Students & Lecturers
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Student
// @Router /students [get]
func (s *StudentService) GetAllStudents(c *fiber.Ctx) error {
    data, err := s.studentRepo.GetAllStudents(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(data)
}

// GetStudentByID godoc
// @Summary Get Student by ID
// @Description Get specific student details
// @Tags Students & Lecturers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student UUID"
// @Success 200 {object} models.Student
// @Failure 404 {object} map[string]interface{}
// @Router /students/{id} [get]
func (s *StudentService) GetStudentByID(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID format"})
    }

    student, err := s.studentRepo.GetStudentByID(c.Context(), id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "student not found"})
    }

    return c.JSON(student)
}

// GetStudentAchievements godoc
// @Summary Get Student Achievements
// @Description Get achievements list for a specific student
// @Tags Students & Lecturers
// @Security BearerAuth
// @Produce json
// @Param id path string true "Student UUID"
// @Success 200 {array} models.Achievement
// @Router /students/{id}/achievements [get]
func (s *StudentService) GetStudentAchievements(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID format"})
    }
    achievements, err := s.achievementRepo.GetStudentAchievements(id)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(achievements)
}

// UpdateAdvisor godoc
// @Summary Update Student Advisor
// @Description Assign or change lecturer advisor for a student
// @Tags Students & Lecturers
// @Security BearerAuth
// @Accept json
// @Param id path string true "Student UUID"
// @Param request body object{lecturerId=string} true "Lecturer UUID"
// @Success 200 {object} map[string]string
// @Router /students/{id}/advisor [put]
func (s *StudentService) UpdateAdvisor(c *fiber.Ctx) error {
    var body struct {
        LecturerID string `json:"lecturerId"`
    }
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
    }

    studentID, err := uuid.Parse(c.Params("id"))
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid Student ID"}) }
    
    lecturerID, err := uuid.Parse(body.LecturerID)
    if err != nil { return c.Status(400).JSON(fiber.Map{"error": "Invalid Lecturer ID"}) }

    err = s.studentRepo.UpdateAdvisor(c.Context(), studentID, lecturerID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{"message": "advisor updated"})
}