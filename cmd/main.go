package main

import (
	"log"

	"github.com/JoniDG/f1-tracker/internal/repository"
	"github.com/JoniDG/f1-tracker/internal/service"
	"github.com/JoniDG/f1-tracker/internal/ui"
)

func main() {
	// Inicializa el repositorio de config (lee/crea ~/.config/f1-tracker/config-f1-tracker.json)
	configRepo, err := repository.NewConfigRepository()
	if err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// Repositorio para consultar info del usuario a la API de Google
	userRepo := repository.NewUserRepository()

	// Crea el servicio de autenticacion, que necesita el configRepo para credenciales
	// y el userRepo para obtener info del usuario despues del login
	authSvc := service.NewAuthService(configRepo, userRepo)

	// Crea el servicio de tracker, que depende de authSvc para tokens validos,
	// configRepo para el spreadsheet ID, y userRepo para info del usuario
	trackerSvc := service.NewTrackerService(authSvc, configRepo, userRepo)

	// Crea y ejecuta la app de Fyne (decide que pantalla mostrar segun el estado de config/token)
	fyneApp := ui.NewFyneApp(configRepo, authSvc, trackerSvc)
	fyneApp.Run()
}
