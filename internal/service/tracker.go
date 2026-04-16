package service

import (
	"log"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/repository"
)

type TrackerService interface {
	GetCurrentUser() (*domain.User, error)
	CreateSpreadsheet(title string) (string, error)
	SaveSpreadsheetID(spreadsheetID string) error
	IsUsernameAvailable(username string) (bool, error)
	SetupUser(username string) error
	NeedsSheetSetup() bool
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

func (s *trackerService) CreateSpreadsheet(title string) (string, error) {
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return "", err
	}
	return s.sheetsRepo.CreateSpreadsheet(token.AccessToken, title)
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
func (s *trackerService) SaveSpreadsheetID(spreadsheetID string) error {
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}
	cfg.SpreadsheetID = spreadsheetID
	err = s.configRepo.SetConfig(*cfg)
	if err != nil {
		return err
	}
	return nil
}
func (s *trackerService) IsUsernameAvailable(userName string) (bool, error) {
	// GetValidToken verifica si el token expiro y lo refresca si es necesario
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return false, err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return false, err
	}
	sheetNames, err := s.GetSheetNames(token.AccessToken, cfg.SpreadsheetID)
	if err != nil {
		return false, err
	}
	if sheetNames[userName] {
		log.Printf("La hoja %s ya existe\n", userName)
		return false, nil
	}
	return true, nil
}
func (s *trackerService) SetupUser(userName string) error {
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}
	cfg.Username = userName
	err = s.configRepo.SetConfig(*cfg)
	if err != nil {
		return err
	}
	err = s.CreateSheet(token.AccessToken, cfg.SpreadsheetID, userName)
	if err != nil {
		return err
	}
	return nil
}
func (s *trackerService) NeedsSheetSetup() bool {
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return true
	}
	return cfg.SpreadsheetID == "" || cfg.Username == ""
}
