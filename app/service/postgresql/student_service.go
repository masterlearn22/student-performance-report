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

func (s *StudentService) GetAllStudents(c *fiber.Ctx) error {
    data, err := s.studentRepo.GetAllStudents(c.Context())
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(data)
}

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