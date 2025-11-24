package service

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPg "student-performance-report/app/repository/postgresql"
)

type ReportService struct {
    mongoRepo   repoMongo.AchievementRepository
    studentRepo repoPg.StudentRepository
}

func NewReportService(m repoMongo.AchievementRepository, s repoPg.StudentRepository) *ReportService {
    return &ReportService{mongoRepo: m, studentRepo: s}
}

func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
    ctx := c.Context()

    stats, err := s.mongoRepo.GetGlobalStats(ctx)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to generate stats"})
    }

    var studentIDs []string
    for _, top := range stats.PointsDistribution {
        studentIDs = append(studentIDs, top.StudentID)
    }

    studentsWithDetails, _ := s.studentRepo.GetStudentsByIDs(ctx, studentIDs) 

    for i, top := range stats.PointsDistribution {
        for _, stud := range studentsWithDetails {
            if top.StudentID == stud.ID.String() {
                stats.PointsDistribution[i].Name = stud.FullName
                stats.PointsDistribution[i].ProgramStudy = stud.ProgramStudy
            }
        }
    }

    return c.JSON(stats)
}

func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
    ctx := c.Context()
    targetStudentID := c.Params("id")

    stats, err := s.mongoRepo.GetStudentStats(ctx, targetStudentID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to get student stats"})
    }

    studentUUID, err := uuid.Parse(targetStudentID)
    if err != nil {
         return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID"})
    }

    studentProfile, err := s.studentRepo.GetStudentByID(ctx, studentUUID) 
    
    if err == nil && studentProfile != nil {
        stats.StudentName = studentProfile.FullName
    }

    return c.JSON(stats)
}
