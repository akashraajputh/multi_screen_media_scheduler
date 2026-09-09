package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"multi-screen-media-scheduler/internal/database"
	"multi-screen-media-scheduler/internal/services"
	"multi-screen-media-scheduler/routes"
	"multi-screen-media-scheduler/seed"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/media_scheduler.db"
	}

	db, err := database.Initialize(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	if err := seed.LoadDefaultData(db); err != nil {
		log.Fatalf("failed to seed default data: %v", err)
	}

	service := services.NewService(db)
	handler := routes.SetupRoutes(service)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := strings.TrimSpace(port)
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}

	fmt.Printf("Server listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
