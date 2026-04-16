package service

import (
	"log"

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
	GetSheet() error
	EnsureSheetExists(sheetName string) error
}

type trackerService struct {
	authSvc    AuthService                 // para obtener un token valido (y refrescarlo si expiro)
	configRepo repository.ConfigRepository // para leer el spreadsheet ID y otros datos de config
	userRepo   repository.UserRepository   // para consultar info del usuario a la API de Google
	sheetsRepo repository.SheetsRepository // para consultar y guardar datos en el spreadsheet
}

func NewTrackerService(authSvc AuthService, configRepo repository.ConfigRepository, userRepo repository.UserRepository, sheetsRepo repository.SheetsRepository) TrackerService {
	return &trackerService{
		authSvc:    authSvc,
		configRepo: configRepo,
		userRepo:   userRepo,
		sheetsRepo: sheetsRepo,
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

func (s *trackerService) GetSheet() error {
	// GetValidToken verifica si el token expiro y lo refresca si es necesario
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}
	return s.sheetsRepo.GetSheetValues(token.AccessToken, cfg.SpreadsheetID, "Hoja 1")
}

func (s *trackerService) EnsureSheetExists(sheetName string) error {
	// GetValidToken verifica si el token expiro y lo refresca si es necesario
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}
	sheetNames, err := s.GetSheetNames(token.AccessToken, cfg.SpreadsheetID)
	if err != nil {
		return err
	}
	if !sheetNames[sheetName] {
		err = s.CreateSheet(token.AccessToken, cfg.SpreadsheetID, sheetName)
		if err != nil {
			return err
		}
	}
	log.Printf("La hoja %s ya existe\n", sheetName)
	return nil
}

func (s *trackerService) GetSheetNames(token, spreadsheetID string) (map[string]bool, error) {
	r, err := s.sheetsRepo.GetSpreadsheetData(token, spreadsheetID)
	if err != nil {
		return nil, err
	}
	sheetNames := make(map[string]bool)
	for _, sheet := range r.Sheets {
		sheetNames[sheet.Properties.Title] = true
	}
	return sheetNames, nil
}
func (s *trackerService) CreateSheet(token, spreadsheetID, userName string) error {
	log.Printf("Se debe crear la hoja %s\n", userName)
	err := s.sheetsRepo.AddSheet(token, spreadsheetID, userName)
	if err != nil {
		return err
	}
	headers := [][]string{{
		"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3",
		"S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha",
	}}
	err = s.sheetsRepo.UpdateSheetValues(token, spreadsheetID, userName+"!A1:J1", headers)
	if err != nil {
		return err
	}
	log.Printf("Hoja %s creada con headers\n", userName)
	return nil
}
