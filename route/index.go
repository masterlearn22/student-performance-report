package route

import (
    "database/sql"
    "github.com/gofiber/fiber/v2"
    repoMongo "student-performance-report/app/repository/mongodb"
    repoPostgre "student-performance-report/app/repository/postgresql"
    mongoService "student-performance-report/app/service/mongodb"
    postgreService "student-performance-report/app/service/postgresql"
    "student-performance-report/database"
    "student-performance-report/middleware"
)

func SetupPostgresRoutes(app *fiber.App, db *sql.DB) {
    // Repositories
    userRepo := repoPostgre.NewUserRepository(db)
    adminRepo := repoPostgre.NewAdminRepository(db)
    studentRepo := repoPostgre.NewStudentRepository(db)
    lecturerRepo := repoPostgre.NewLecturerRepository(db)
    achRepoPg := repoPostgre.NewAchievementRepoPostgres(db)
    achRepoMongo := repoMongo.NewAchievementRepository(database.MongoDB)

    // Services
    authService := postgreService.NewAuthService(userRepo)
    adminService := postgreService.NewAdminService(adminRepo, userRepo)
    lecturerService := postgreService.NewLecturerService(lecturerRepo)
    studentService := postgreService.NewStudentService(studentRepo, achRepoMongo)
    achievementService := mongoService.NewAchievementService(achRepoMongo, achRepoPg, lecturerRepo)
	reportService := mongoService.NewReportService(achRepoMongo, studentRepo)

    // Static Files Config
    app.Static("/uploads", "./uploads")   
    api := app.Group("/api/v1")

    // 5.1 Authentication
    auth := api.Group("/auth")
    auth.Post("/login", authService.Login)
    auth.Post("/refresh", authService.Refresh)
    auth.Post("/logout", middleware.AuthRequired(), authService.Logout)
    auth.Get("/profile", middleware.AuthRequired(), authService.Profile)

    // 5.2 Users 
    users := api.Group("/users", middleware.AuthRequired())
    users.Get("/", adminService.GetAllUsers)
    users.Get("/:id", adminService.GetUserByID)
    users.Post("/", adminService.CreateUser)
    users.Put("/:id", adminService.UpdateUser)
    users.Delete("/:id", adminService.DeleteUser)
    users.Put("/:id/role", adminService.AssignRole)

    // 5.4 Achievements
    ach := api.Group("/achievements", middleware.AuthRequired())
    ach.Get("/", achievementService.GetAllAchievements)
    ach.Get("/:id", achievementService.GetAchievementDetail)
    ach.Get("/:id/history", achievementService.GetAchievementHistory)
    ach.Post("/", achievementService.CreateAchievement) 
    ach.Put("/:id", achievementService.UpdateAchievement)
    ach.Delete("/:id",  achievementService.DeleteAchievement)
    ach.Post("/:id/submit", achievementService.SubmitAchievement)
    ach.Post("/:id/attachments", achievementService.UploadAttachments)
    ach.Post("/:id/verify", achievementService.VerifyAchievement)
    ach.Post("/:id/reject", achievementService.RejectAchievement)

    // 5.5 Students & Lecturers
    student := api.Group("/students", middleware.AuthRequired())
    lecturer := api.Group("/lecturers", middleware.AuthRequired())
    student.Get("/",  studentService.GetAllStudents)
    student.Get("/:id",   studentService.GetStudentByID)
    student.Get("/:id/achievements", studentService.GetStudentAchievements)
    student.Put("/:id/advisor",   studentService.UpdateAdvisor)
    lecturer.Get("/",   lecturerService.GetAllLecturers)
    lecturer.Get("/:id/advisees", lecturerService.GetAdvisees)

	// 5.8 Reports & Analytics (NEW)
	reports := api.Group("/reports", middleware.AuthRequired())    
	reports.Get("/statistics", reportService.GetStatistics)
	reports.Get("/student/:id", reportService.GetStudentReport)
}