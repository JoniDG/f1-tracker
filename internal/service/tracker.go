package service

import (
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/repository"
)

type TrackerService interface {
	// TODO: Implementar cuando SheetsRepository este listo
	// GetMyTracks() ([]domain.TrackTime, error)
	// SaveTrackTime(track domain.TrackTime) error
	// GetFriendsList() ([]string, error)
	// GetFriendTracks(name string) ([]domain.TrackTime, error)
	// EnsureSheetExists() error

	// GetCurrentUser devuelve la info del usuario logueado usando un token valido.
	// Si el token expiro, lo refresca automaticamente via AuthService.
	GetCurrentUser() (*domain.User, error)
}

type trackerService struct {
	authSvc    AuthService                 // para obtener un token valido (y refrescarlo si expiro)
	configRepo repository.ConfigRepository // para leer el spreadsheet ID y otros datos de config
	userRepo   repository.UserRepository   // para consultar info del usuario a la API de Google
	// TODO: Agregar sheetsRepo repository.SheetsRepository cuando se implemente
}

func NewTrackerService(authSvc AuthService, configRepo repository.ConfigRepository, userRepo repository.UserRepository) TrackerService {
	return &trackerService{
		authSvc:    authSvc,
		configRepo: configRepo,
		userRepo:   userRepo,
	}
}

func (s *trackerService) GetCurrentUser() (*domain.User, error) {
	// GetValidToken verifica si el token expiro y lo refresca si es necesario
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return nil, err
	}

	// Consulta la API de Google con el token valido para obtener nombre y email
	return s.userRepo.GetUserInfo(token.AccessToken)
}

// TODO: Implementar los metodos de TrackerService cuando SheetsRepository este listo.
// Cada metodo deberia:
// 1. Obtener un token valido con s.authSvc.GetValidToken()
// 2. Leer el spreadsheet ID con s.configRepo.GetConfig()
// 3. Llamar al metodo correspondiente de s.sheetsRepo
//
// Ejemplo:
// func (s *trackerService) GetMyTracks() ([]domain.TrackTime, error) {
//     token, err := s.authSvc.GetValidToken()
//     cfg, err := s.configRepo.GetConfig()
//     return s.sheetsRepo.GetAllTrackTimes(token.AccessToken, cfg.SpreadsheetID, userName)
// }
