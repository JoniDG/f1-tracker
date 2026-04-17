package main

import (
	"log"

	"github.com/JoniDG/f1-tracker/internal/repository"
	"github.com/JoniDG/f1-tracker/internal/service"
	"github.com/JoniDG/f1-tracker/internal/ui"
)

func main() {
	configRepo, err := repository.NewConfigRepository()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	userRepo := repository.NewUserRepository()
	sheetsRepo := repository.NewSheetsRepository()
	authSvc := service.NewAuthService(configRepo, userRepo)
	trackerSvc := service.NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	fyneApp := ui.NewFyneApp(authSvc, trackerSvc)
	fyneApp.Run()
}
