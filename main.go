package main

import (
	"fmt"
	"os"
	"log"
	"os/signal"
	"syscall"
	"time"
	"context"
	"student-performance-report/config"
	"student-performance-report/database"
	FiberApp "student-performance-report/fiber"
	routePostgre "student-performance-report/route/postgresql"
	routeMongo "student-performance-report/route/mongodb"

	
)

func main() {

	// 1. Load .env file
    config.LoadEnv() // Load file .env
    // host := os.Getenv("DB_HOST")
    // if host == "" {
    //     fmt.Println(".env gagal diload atau DB_HOST tidak ditemukan")
    // } else {
    //     fmt.Println(".env berhasil diload. DB_HOST =", host)
    // }

	//2. Connect to Database

	// Connect to PostgreSQL
	database.ConnectPostgres()
	defer database.PostgresDB.Close()

	// Connect to MongoDB
	database.ConnectMongo()

	//3 Setup Fiber App
	app := FiberApp.SetupFiber()

	//4. Setup Route
	routePostgre.SetupPostgresRoutes(app, database.PostgresDB)
	routeMongo.SetupMongoRoutes(app, database.MongoDB)

	fmt.Println("Setup route berhasil")

	// 5 Start server
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

	// 6 Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
