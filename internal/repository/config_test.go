package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JoniDG/f1-tracker/internal/defines"
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func newTestConfigRepo(t *testing.T) ConfigRepository {
	t.Helper()
	return newTestConfigRepoAt(t, filepath.Join(t.TempDir(), defines.ConfigPath))
}

func newTestConfigRepoAt(t *testing.T, configPath string) ConfigRepository {
	t.Helper()
	require.NoError(t, os.MkdirAll(configPath, 0o750))

	v := viper.New()
	v.SetConfigType("json")
	v.AddConfigPath(configPath)
	v.SetConfigName(defines.ConfigFilename)

	cr := &configRepository{
		viper:        v,
		configLoaded: true,
	}

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			require.NoError(t, err)
		}
		require.NoError(t, v.SafeWriteConfig())
		cr.configLoaded = false
	}

	return cr
}

func TestNewConfigRepository_WhenNewDir_ShouldCreateConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)

	_ = newTestConfigRepoAt(t, configPath)

	assert.FileExists(t, filepath.Join(configPath, defines.ConfigFilename+".json"))
}

func TestNewConfigRepository_WhenExistingConfig_ShouldLoadSuccessfully(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)
	require.NoError(t, os.MkdirAll(configPath, 0o750))
	configFile := filepath.Join(configPath, defines.ConfigFilename+".json")
	require.NoError(t, os.WriteFile(configFile, []byte(`{"config":{"googleclientid":"test-id"}}`), 0o600))

	repo := newTestConfigRepoAt(t, configPath)

	assert.NotNil(t, repo)
}

func TestNewConfigRepository_WhenInvalidPath_ShouldReturnError(t *testing.T) {
	repo, err := NewConfigRepository()

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

func TestConfigRepository_HasValidConfig_WhenValidOAuthCredentials_ShouldReturnTrue(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{
		GoogleClientID:     "test-id.apps.googleusercontent.com",
		GoogleClientSecret: "secret-123",
	}))

	assert.True(t, repo.HasValidConfig())
}

func TestConfigRepository_HasValidConfig_WithoutSpreadsheetID_ShouldStillReturnTrue(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{
		GoogleClientID:     "test-id.apps.googleusercontent.com",
		GoogleClientSecret: "secret-123",
		CallbackPort:       "8081",
	}))

	assert.True(t, repo.HasValidConfig())
}

func TestConfigRepository_HasValidConfig_WhenMissingClientID_ShouldReturnFalse(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{
		GoogleClientSecret: "secret-123",
	}))

	assert.False(t, repo.HasValidConfig())
}

func TestConfigRepository_HasValidConfig_WhenInvalidClientIDSuffix_ShouldReturnFalse(t *testing.T) {
	repo := newTestConfigRepo(t)
	require.NoError(t, repo.SetConfig(domain.Config{
		GoogleClientID:     "test-id-invalid",
		GoogleClientSecret: "secret-123",
	}))

	assert.False(t, repo.HasValidConfig())
}

func TestConfigRepository_HasValidConfig_WhenNotLoaded_ShouldReturnFalse(t *testing.T) {
	repo := newTestConfigRepo(t)

	assert.False(t, repo.HasValidConfig())
}

func TestConfigRepository_GetConfig_WhenPersisted_ShouldSurviveNewInstance(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, defines.ConfigPath)

	repo1 := newTestConfigRepoAt(t, configPath)
	expected := domain.Config{
		GoogleClientID:     "persist-id",
		GoogleClientSecret: "persist-secret",
		CallbackPort:       "7777",
		SpreadsheetID:      "persist-sheet",
	}
	require.NoError(t, repo1.SetConfig(expected))

	repo2 := newTestConfigRepoAt(t, configPath)

	got, err := repo2.GetConfig()
	require.NoError(t, err)
	assert.Equal(t, expected.GoogleClientID, got.GoogleClientID)
	assert.Equal(t, expected.SpreadsheetID, got.SpreadsheetID)
}
