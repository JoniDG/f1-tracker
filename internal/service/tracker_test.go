package service

import (
	"errors"
	"testing"

	"github.com/JoniDG/f1-tracker/internal/defines"
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
	rows := [][]string{{
		"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3",
		"S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha",
	}}
	for _, track := range defines.Tracks {
		rows = append(rows, []string{track, "", "", "", "", "", "", "", "", ""})
	}
	sheetsRepo.On("UpdateSheetValues", "valid-token", "sheet-123", "JoniDG!A1:J25", rows).Return(nil)

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

func TestTrackerService_GetMyTracks_WhenSuccess_ShouldReturnTracks(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	values := [][]string{
		{"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3", "S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha"},
		{"Bahrain", "1:23.456", "0:28.100", "0:27.200", "0:28.156", "0:28.100", "0:27.200", "0:28.156", "Ferrari", "2026-04-09"},
		{"Saudi Arabia"},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSheetValues", "valid-token", "sheet-123", "JoniDG").Return(values, nil)

	tracks, err := svc.GetMyTracks()

	require.NoError(t, err)
	assert.Len(t, tracks, 2)
	assert.Equal(t, "Bahrain", tracks[0].TrackName)
	assert.Equal(t, "1:23.456", tracks[0].BestLapTime)
	assert.Equal(t, "Ferrari", tracks[0].Car)
	assert.Equal(t, "Saudi Arabia", tracks[1].TrackName)
	assert.Empty(t, tracks[1].BestLapTime)
}

func TestTrackerService_GetMyTracks_WhenOnlyHeaders_ShouldReturnEmpty(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	values := [][]string{
		{"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3", "S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha"},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSheetValues", "valid-token", "sheet-123", "JoniDG").Return(values, nil)

	tracks, err := svc.GetMyTracks()

	require.NoError(t, err)
	assert.Empty(t, tracks)
}

func TestTrackerService_GetMyTracks_WhenTokenError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	authSvc.On("GetValidToken").Return(nil, errors.New("token expired"))

	tracks, err := svc.GetMyTracks()

	assert.Nil(t, tracks)
	assert.EqualError(t, err, "token expired")
}

func TestTrackerService_SaveTrackTime_WhenSuccess_ShouldWriteRow(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)

	track := domain.TrackTime{
		TrackName: "Bahrain", BestLapTime: "1:23.456",
		BestS1: "0:28.1", BestS2: "0:27.2", BestS3: "0:28.1",
		LapS1: "0:28.1", LapS2: "0:27.2", LapS3: "0:28.1",
		Car: "Ferrari", Date: "2026-04-09",
	}
	expectedRow := [][]string{{
		"Bahrain", "1:23.456", "0:28.1", "0:27.2", "0:28.1",
		"0:28.1", "0:27.2", "0:28.1", "Ferrari", "2026-04-09",
	}}

	// Bahrain es indice 0 en defines.Tracks -> fila 2
	sheetsRepo.On("UpdateSheetValues", "valid-token", "sheet-123", "JoniDG!A2:J2", expectedRow).Return(nil)

	err := svc.SaveTrackTime(track)

	require.NoError(t, err)
	sheetsRepo.AssertExpectations(t)
}

func TestTrackerService_SaveTrackTime_WhenInvalidTrack_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	track := domain.TrackTime{TrackName: "Circuito Inexistente"}

	err := svc.SaveTrackTime(track)

	assert.ErrorContains(t, err, "circuito no encontrado")
}

func TestTrackerService_SaveTrackTime_WhenLastTrack_ShouldUseCorrectRow(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)

	track := domain.TrackTime{TrackName: "Abu Dhabi", BestLapTime: "1:30.000"}
	expectedRow := [][]string{{"Abu Dhabi", "1:30.000", "", "", "", "", "", "", "", ""}}

	// Abu Dhabi es indice 23 en defines.Tracks -> fila 25
	sheetsRepo.On("UpdateSheetValues", "valid-token", "sheet-123", "JoniDG!A25:J25", expectedRow).Return(nil)

	err := svc.SaveTrackTime(track)

	require.NoError(t, err)
	sheetsRepo.AssertExpectations(t)
}

func TestTrackerService_GetFriendsList_WhenFriendsExist_ShouldReturnFiltered(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	spreadsheetData := &domain.SpreadsheetData{
		Sheets: []domain.SheetData{
			{Properties: domain.SheetDataProperties{Title: "JoniDG"}},
			{Properties: domain.SheetDataProperties{Title: "Amigo1"}},
			{Properties: domain.SheetDataProperties{Title: "Amigo2"}},
		},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSpreadsheetData", "valid-token", "sheet-123").Return(spreadsheetData, nil)

	friends, err := svc.GetFriendsList()

	require.NoError(t, err)
	assert.Len(t, friends, 2)
	assert.Contains(t, friends, "Amigo1")
	assert.Contains(t, friends, "Amigo2")
	assert.NotContains(t, friends, "JoniDG")
}

func TestTrackerService_GetFriendsList_WhenNoFriends_ShouldReturnEmpty(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	spreadsheetData := &domain.SpreadsheetData{
		Sheets: []domain.SheetData{
			{Properties: domain.SheetDataProperties{Title: "JoniDG"}},
		},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSpreadsheetData", "valid-token", "sheet-123").Return(spreadsheetData, nil)

	friends, err := svc.GetFriendsList()

	require.NoError(t, err)
	assert.Empty(t, friends)
}

func TestTrackerService_GetFriendTracks_WhenSuccess_ShouldReturnTracks(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	values := [][]string{
		{"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3", "S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha"},
		{"Bahrain", "1:25.000", "", "", "", "", "", "", "McLaren", "2026-04-10"},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSheetValues", "valid-token", "sheet-123", "Amigo1").Return(values, nil)

	tracks, err := svc.GetFriendTracks("Amigo1")

	require.NoError(t, err)
	assert.Len(t, tracks, 1)
	assert.Equal(t, "Bahrain", tracks[0].TrackName)
	assert.Equal(t, "McLaren", tracks[0].Car)
}

func TestTrackerService_GetFriendTracks_WhenOnlyHeaders_ShouldReturnEmpty(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}
	values := [][]string{
		{"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3", "S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha"},
	}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSheetValues", "valid-token", "sheet-123", "AmigoSinDatos").Return(values, nil)

	tracks, err := svc.GetFriendTracks("AmigoSinDatos")

	require.NoError(t, err)
	assert.Empty(t, tracks)
}

func TestTrackerService_GetFriendTracks_WhenError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	sheetsRepo := new(mocks.MockSheetsRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo, sheetsRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	cfg := &domain.Config{SpreadsheetID: "sheet-123", Username: "JoniDG"}

	authSvc.On("GetValidToken").Return(token, nil)
	configRepo.On("GetConfig").Return(cfg, nil)
	sheetsRepo.On("GetSheetValues", "valid-token", "sheet-123", "Amigo1").Return(nil, errors.New("API error"))

	tracks, err := svc.GetFriendTracks("Amigo1")

	assert.Nil(t, tracks)
	assert.EqualError(t, err, "API error")
}
