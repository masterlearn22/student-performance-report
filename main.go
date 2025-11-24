package main

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @tag.name Authentication
// @tag.description Endpoint for user authentication and login
// @tag.order 1

// @tag.name Users 
// @tag.description Endpoint for Admin to manage users (Admin Only)
// @tag.order 2

// @tag.name Achievements
// @tag.description Endpoint for achievement data
// @tag.order 3

// @tag.name Students & Lecturers
// @tag.description Endpoint for student and lecturer data 
// @tag.order 4

// @tag.name Reports
// @tag.description Endpoint for generating reports and statistics
// @tag.order 5

import (
	"fmt"
	"os"
	"log"
	"os/signal"
	"syscall"
	"time"
	"context"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"student-performance-report/config"
	"student-performance-report/database"
	FiberApp "student-performance-report/fiber"
	route "student-performance-report/route"
	"github.com/gofiber/swagger"
	docs "student-performance-report/docs"
)

// @title Student Performance Report API
// @version 1.0
// @description API untuk sistem pelaporan prestasi mahasiswa.
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	// 1. Load .env file
    config.LoadEnv() 

	// 2. Connect to Database
	// Connect to PostgreSQL
	database.ConnectPostgres()
	defer database.PostgresDB.Close()

	// Connect to MongoDB
	database.ConnectMongo()

	// 3. Setup Fiber App
	app := FiberApp.SetupFiber()
	app.Use(logger.New())

	// 4. Swagger
	docs.SwaggerInfo.BasePath = "/api/v1"
	app.Get("/swagger/*", swagger.HandlerDefault) 
	log.Println("➡️  Swagger UI available at: http://localhost:" + os.Getenv("PORT") + "/swagger/index.html")

	// 5. Setup Route
	route.SetupPostgresRoutes(app, database.PostgresDB)

	fmt.Println("Setup route berhasil")

	// 6. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	go func() {
		log.Printf("Server running on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server stopped: %v", err)
		}
	}()

	// 7. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
