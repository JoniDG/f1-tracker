package service

import (
	"errors"
	"testing"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewTrackerService_ShouldReturnInstance(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)

	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	assert.NotNil(t, svc)
}

func TestTrackerService_GetCurrentUser_WhenValidToken_ShouldReturnUser(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	expectedUser := &domain.User{DisplayName: "Juan", Email: "juan@test.com"}

	authSvc.On("GetValidToken").Return(token, nil)
	userRepo.On("GetUserInfo", "valid-token").Return(expectedUser, nil)

	user, err := svc.GetCurrentUser()

	require.NoError(t, err)
	assert.Equal(t, "Juan", user.DisplayName)
	assert.Equal(t, "juan@test.com", user.Email)
	authSvc.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestTrackerService_GetCurrentUser_WhenTokenError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	authSvc.On("GetValidToken").Return(nil, errors.New("token expired"))

	user, err := svc.GetCurrentUser()

	assert.Nil(t, user)
	assert.EqualError(t, err, "token expired")
	authSvc.AssertExpectations(t)
}

func TestTrackerService_GetCurrentUser_WhenUserInfoError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	authSvc.On("GetValidToken").Return(token, nil)
	userRepo.On("GetUserInfo", "valid-token").Return(nil, errors.New("API error"))

	user, err := svc.GetCurrentUser()

	assert.Nil(t, user)
	assert.EqualError(t, err, "API error")
	authSvc.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestTrackerService_CreateSpreadsheet_WhenSuccess_ShouldReturnID(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	authSvc.On("GetValidToken").Return(token, nil)
	sheetsRepo.On("CreateSpreadsheet", "valid-token", "F1 Tiempos").Return("spreadsheet-123", nil)

	id, err := svc.CreateSpreadsheet("F1 Tiempos")

	require.NoError(t, err)
	assert.Equal(t, "spreadsheet-123", id)
	authSvc.AssertExpectations(t)
	sheetsRepo.AssertExpectations(t)
}

func TestTrackerService_CreateSpreadsheet_WhenTokenError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	authSvc.On("GetValidToken").Return(nil, errors.New("token expired"))

	id, err := svc.CreateSpreadsheet("F1 Tiempos")

	assert.Empty(t, id)
	assert.EqualError(t, err, "token expired")
}

func TestTrackerService_SaveSpreadsheetID_WhenSuccess_ShouldPersist(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	cfg := &domain.Config{GoogleClientID: "test-id"}
	configRepo.On("GetConfig").Return(cfg, nil)
	configRepo.On("SetConfig", domain.Config{GoogleClientID: "test-id", SpreadsheetID: "sheet-456"}).Return(nil)

	err := svc.SaveSpreadsheetID("sheet-456")

	require.NoError(t, err)
	configRepo.AssertExpectations(t)
}

func TestTrackerService_SaveSpreadsheetID_WhenGetConfigError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	configRepo.On("GetConfig").Return(nil, errors.New("config error"))

	err := svc.SaveSpreadsheetID("sheet-456")

	assert.EqualError(t, err, "config error")
}

func TestTrackerService_IsUsernameAvailable_WhenAvailable_ShouldReturnTrue(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123"}
	spreadsheetData := &domain.SpreadsheetData{
		Sheets: []domain.SheetData{
			{Properties: domain.SheetDataProperties{Title: "OtroUsuario"}},
		},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSpreadsheetData", "valid-token", "sheet-123").Return(spreadsheetData, nil)

	available, err := svc.IsUsernameAvailable("NuevoUsuario")

	require.NoError(t, err)
	assert.True(t, available)
}

func TestTrackerService_IsUsernameAvailable_WhenTaken_ShouldReturnFalse(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123"}
	spreadsheetData := &domain.SpreadsheetData{
		Sheets: []domain.SheetData{
			{Properties: domain.SheetDataProperties{Title: "JoniDG"}},
		},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSpreadsheetData", "valid-token", "sheet-123").Return(spreadsheetData, nil)

	available, err := svc.IsUsernameAvailable("JoniDG")

	require.NoError(t, err)
	assert.False(t, available)
}

func TestTrackerService_SetupUser_WhenSuccess_ShouldSaveConfigAndCreateSheet(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{GoogleClientID: "test-id", SpreadsheetID: "sheet-123"}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	configRepo.On("SetConfig", domain.Config{GoogleClientID: "test-id", SpreadsheetID: "sheet-123", Username: "JoniDG"}).Return(nil)
	sheetsRepo.On("AddSheet", "valid-token", "sheet-123", "JoniDG").Return(nil)
	headers := [][]string{{
		"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3",
		"S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha",
	}}
	sheetsRepo.On("UpdateSheetValues", "valid-token", "sheet-123", "JoniDG!A1:J1", headers).Return(nil)

	err := svc.SetupUser("JoniDG")

	require.NoError(t, err)
	configRepo.AssertExpectations(t)
	sheetsRepo.AssertExpectations(t)
}

func TestTrackerService_NeedsSheetSetup_WhenMissingSpreadsheetID_ShouldReturnTrue(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	cfg := &domain.Config{Username: "JoniDG"}
	configRepo.On("GetConfig").Return(cfg, nil)

	assert.True(t, svc.NeedsSheetSetup())
}

func TestTrackerService_NeedsSheetSetup_WhenMissingUsername_ShouldReturnTrue(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	cfg := &domain.Config{SpreadsheetID: "sheet-123"}
	configRepo.On("GetConfig").Return(cfg, nil)

	assert.True(t, svc.NeedsSheetSetup())
}

func TestTrackerService_NeedsSheetSetup_WhenComplete_ShouldReturnFalse(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	configRepo.On("GetConfig").Return(cfg, nil)

	assert.False(t, svc.NeedsSheetSetup())
}

func TestTrackerService_NeedsSheetSetup_WhenConfigError_ShouldReturnTrue(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	configRepo.On("GetConfig").Return(nil, errors.New("config error"))

	assert.True(t, svc.NeedsSheetSetup())
}
