package service

import (
	"fmt"
	"log"

	"github.com/JoniDG/f1-tracker/internal/defines"
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/repository"
)

type TrackerService interface {
	GetCurrentUser() (*domain.User, error)
	CreateSpreadsheet(title string) (string, error)
	SaveSpreadsheetID(spreadsheetID string) error
	IsUsernameAvailable(username string) (bool, error)
	SetupUser(username string, cleanupDefault bool) error
	NeedsSheetSetup() bool
	GetMyTracks() ([]domain.TrackTime, error)
	SaveTrackTime(track domain.TrackTime) error
	GetFriendsList() ([]string, error)
	GetFriendTracks(friendName string) ([]domain.TrackTime, error)
}

type trackerService struct {
	authSvc    AuthService
	configRepo repository.ConfigRepository
	userRepo   repository.UserRepository
	sheetsRepo repository.SheetsRepository
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
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return nil, err
	}

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

	rows := [][]string{{
		"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3",
		"S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha",
	}}
	for _, track := range defines.Tracks {
		rows = append(rows, []string{track, "", "", "", "", "", "", "", "", ""})
	}

	rangeStr := fmt.Sprintf("%s!A1:J%d", userName, len(rows))
	err = s.sheetsRepo.UpdateSheetValues(token, spreadsheetID, rangeStr, rows)
	if err != nil {
		return err
	}
	log.Printf("Hoja %s creada con headers y %d circuitos\n", userName, len(defines.Tracks))
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

func (s *trackerService) SetupUser(userName string, cleanupDefault bool) error {
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
	if cleanupDefault {
		if err := s.deleteDefaultSheet(token.AccessToken, cfg.SpreadsheetID, userName); err != nil {
			log.Printf("no se pudo eliminar la hoja por defecto: %v", err)
		}
	}
	return nil
}

func (s *trackerService) deleteDefaultSheet(token, spreadsheetID, userName string) error {
	data, err := s.sheetsRepo.GetSpreadsheetData(token, spreadsheetID)
	if err != nil {
		return err
	}
	for _, sheet := range data.Sheets {
		if sheet.Properties.SheetId == 0 && sheet.Properties.Title != userName {
			return s.sheetsRepo.DeleteSheet(token, spreadsheetID, sheet.Properties.SheetId)
		}
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

func (s *trackerService) GetMyTracks() ([]domain.TrackTime, error) {
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return nil, err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, err
	}
	values, err := s.sheetsRepo.GetSheetValues(token.AccessToken, cfg.SpreadsheetID, cfg.Username)
	if err != nil {
		return nil, err
	}
	return parseRows(values), nil
}

func (s *trackerService) SaveTrackTime(track domain.TrackTime) error {
	trackIndex := -1
	for i, name := range defines.Tracks {
		if name == track.TrackName {
			trackIndex = i
			break
		}
	}
	if trackIndex == -1 {
		return fmt.Errorf("circuito no encontrado: %s", track.TrackName)
	}

	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return err
	}

	row := trackIndex + 2
	rangeStr := fmt.Sprintf("%s!A%d:J%d", cfg.Username, row, row)
	values := [][]string{trackTimeToRow(track)}

	return s.sheetsRepo.UpdateSheetValues(token.AccessToken, cfg.SpreadsheetID, rangeStr, values)
}

func (s *trackerService) GetFriendsList() ([]string, error) {
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return nil, err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, err
	}
	sheetNames, err := s.GetSheetNames(token.AccessToken, cfg.SpreadsheetID)
	if err != nil {
		return nil, err
	}

	var friends []string
	for name := range sheetNames {
		if name != cfg.Username {
			friends = append(friends, name)
		}
	}
	return friends, nil
}

func (s *trackerService) GetFriendTracks(friendName string) ([]domain.TrackTime, error) {
	token, err := s.authSvc.GetValidToken()
	if err != nil {
		return nil, err
	}
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, err
	}
	values, err := s.sheetsRepo.GetSheetValues(token.AccessToken, cfg.SpreadsheetID, friendName)
	if err != nil {
		return nil, err
	}
	return parseRows(values), nil
}

func parseRows(values [][]string) []domain.TrackTime {
	var tracks []domain.TrackTime
	for i, row := range values {
		if i == 0 {
			continue
		}
		if len(row) == 0 {
			continue
		}
		track := domain.TrackTime{TrackName: safeGet(row, 0)}
		track.BestLapTime = safeGet(row, 1)
		track.BestS1 = safeGet(row, 2)
		track.BestS2 = safeGet(row, 3)
		track.BestS3 = safeGet(row, 4)
		track.LapS1 = safeGet(row, 5)
		track.LapS2 = safeGet(row, 6)
		track.LapS3 = safeGet(row, 7)
		track.Car = safeGet(row, 8)
		track.Date = safeGet(row, 9)
		tracks = append(tracks, track)
	}
	return tracks
}

func safeGet(row []string, index int) string {
	if index < len(row) {
		return row[index]
	}
	return ""
}

func trackTimeToRow(track domain.TrackTime) []string {
	return []string{
		track.TrackName,
		track.BestLapTime,
		track.BestS1,
		track.BestS2,
		track.BestS3,
		track.LapS1,
		track.LapS2,
		track.LapS3,
		track.Car,
		track.Date,
	}
}
