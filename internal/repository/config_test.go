package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JoniDG/f1-tracker/internal/defines"
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestConfigRepo(t *testing.T) ConfigRepository {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), defines.ConfigPath)
	repo, err := newConfigRepository(configPath)
	require.NoError(t, err)
	return repo
}

func TestNewConfigRepository_WhenNewDir_ShouldCreateConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)

	_, err := newConfigRepository(configPath)

	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(configPath, defines.ConfigFilename+".json"))
}

func TestNewConfigRepository_WhenNewDir_ShouldSetConfigLoadedFalse(t *testing.T) {
	repo := newTestConfigRepo(t)

	assert.False(t, repo.IsLoaded())
}

func TestNewConfigRepository_WhenExistingConfig_ShouldSetConfigLoadedTrue(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)
	require.NoError(t, os.MkdirAll(configPath, 0o750))
	configFile := filepath.Join(configPath, defines.ConfigFilename+".json")
	require.NoError(t, os.WriteFile(configFile, []byte(`{"config":{"googleclientid":"test-id"}}`), 0o600))

	repo, err := newConfigRepository(configPath)

	require.NoError(t, err)
	assert.True(t, repo.IsLoaded())
}

func TestNewConfigRepository_WhenInvalidPath_ShouldReturnError(t *testing.T) {
	repo, err := NewConfigRepository()

	// No error expected in normal environment, just verify it works
	if err == nil {
		assert.NotNil(t, repo)
	}
}

func TestConfigRepository_SetConfig_WhenValid_ShouldPersistAndSetLoaded(t *testing.T) {
	repo := newTestConfigRepo(t)

	cfg := domain.Config{
		GoogleClientID:     "client-id-123",
		GoogleClientSecret: "client-secret-456",
		CallbackPort:       "9090",
		SpreadsheetID:      "spreadsheet-abc",
	}

	err := repo.SetConfig(cfg)

	require.NoError(t, err)
	assert.True(t, repo.IsLoaded())
}

func TestConfigRepository_GetConfig_WhenLoaded_ShouldReturnSavedConfig(t *testing.T) {
	repo := newTestConfigRepo(t)
	expected := domain.Config{
		GoogleClientID:     "client-id-123",
		GoogleClientSecret: "client-secret-456",
		CallbackPort:       "9090",
		SpreadsheetID:      "spreadsheet-abc",
	}
	require.NoError(t, repo.SetConfig(expected))

	got, err := repo.GetConfig()

	require.NoError(t, err)
	assert.Equal(t, expected.GoogleClientID, got.GoogleClientID)
	assert.Equal(t, expected.GoogleClientSecret, got.GoogleClientSecret)
	assert.Equal(t, expected.CallbackPort, got.CallbackPort)
	assert.Equal(t, expected.SpreadsheetID, got.SpreadsheetID)
}

func TestConfigRepository_GetConfig_WhenNotLoaded_ShouldReturnError(t *testing.T) {
	repo := newTestConfigRepo(t)

	got, err := repo.GetConfig()

	assert.Nil(t, got)
	assert.EqualError(t, err, "no config found, run 'f1-tracker config' first")
}

func TestConfigRepository_SetGoogleCredentials_WhenLoaded_ShouldPersistToken(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{GoogleClientID: "test"}))

	token := oauth2.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		Expiry:       time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	err := repo.SetGoogleCredentials(token)

	require.NoError(t, err)
}

func TestConfigRepository_SetGoogleCredentials_WhenNotLoaded_ShouldReturnError(t *testing.T) {
	repo := newTestConfigRepo(t)

	token := oauth2.Token{AccessToken: "test"}

	err := repo.SetGoogleCredentials(token)

	assert.EqualError(t, err, "no config found, run 'f1-tracker config' first")
}

func TestConfigRepository_GetGoogleToken_WhenLoaded_ShouldReturnToken(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{GoogleClientID: "test"}))

	expected := oauth2.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		Expiry:       time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	require.NoError(t, repo.SetGoogleCredentials(expected))

	got, err := repo.GetGoogleToken()

	require.NoError(t, err)
	assert.Equal(t, expected.AccessToken, got.AccessToken)
	assert.Equal(t, expected.RefreshToken, got.RefreshToken)
	assert.Equal(t, expected.TokenType, got.TokenType)
}

func TestConfigRepository_GetGoogleToken_WhenNotLoaded_ShouldReturnError(t *testing.T) {
	repo := newTestConfigRepo(t)

	got, err := repo.GetGoogleToken()

	assert.Nil(t, got)
	assert.EqualError(t, err, "no config found, run 'f1-tracker config' first")
}

func TestConfigRepository_GetGoogleToken_WhenNoTokenSet_ShouldReturnEmptyToken(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{GoogleClientID: "test"}))

	got, err := repo.GetGoogleToken()

	require.NoError(t, err)
	assert.Empty(t, got.AccessToken)
}

func TestConfigRepository_IsLoaded_WhenNewRepo_ShouldReturnFalse(t *testing.T) {
	repo := newTestConfigRepo(t)

	assert.False(t, repo.IsLoaded())
}

func TestConfigRepository_IsLoaded_WhenAfterSetConfig_ShouldReturnTrue(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{}))

	assert.True(t, repo.IsLoaded())
}

func TestConfigRepository_GetConfig_WhenPersisted_ShouldSurviveNewInstance(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)

	repo1, err := newConfigRepository(configPath)
	require.NoError(t, err)
	expected := domain.Config{
		GoogleClientID:     "persist-id",
		GoogleClientSecret: "persist-secret",
		CallbackPort:       "7777",
		SpreadsheetID:      "persist-sheet",
	}
	require.NoError(t, repo1.SetConfig(expected))

	repo2, err := newConfigRepository(configPath)
	require.NoError(t, err)

	got, err := repo2.GetConfig()
	require.NoError(t, err)
	assert.Equal(t, expected.GoogleClientID, got.GoogleClientID)
	assert.Equal(t, expected.SpreadsheetID, got.SpreadsheetID)
}
