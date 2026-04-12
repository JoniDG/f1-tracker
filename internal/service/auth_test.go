package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// saveAndRestoreGlobals saves the package-level vars and restores them after the test.
func saveAndRestoreGlobals(t *testing.T) {
	t.Helper()
	origBrowserOpen := browserOpenFunc
	origEndpoint := googleEndpoint
	t.Cleanup(func() {
		browserOpenFunc = origBrowserOpen
		googleEndpoint = origEndpoint
	})
}

func defaultTestConfig() *domain.Config {
	return &domain.Config{
		GoogleClientID:     "test-client-id",
		GoogleClientSecret: "test-client-secret",
		CallbackPort:       "0",
		SpreadsheetID:      "test-sheet",
	}
}

// fakeTokenServer creates an httptest server that returns a valid OAuth token response.
func fakeTokenServer(t *testing.T, accessToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]interface{}{
			"access_token":  accessToken,
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "refresh-" + accessToken,
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(resp)
		require.NoError(t, err)
	}))
}

// fakeErrorTokenServer creates an httptest server that returns an OAuth error.
func fakeErrorTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		require.NoError(t, err)
	}))
}

// simulateCallback sends an HTTP GET to the local callback server and asserts success.
func simulateCallback(t *testing.T, port, query string, expectedStatus int) {
	t.Helper()
	require.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:" + port + "/callback?" + query)
		if err != nil {
			return false
		}
		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
		return resp.StatusCode == expectedStatus
	}, 3*time.Second, 50*time.Millisecond)
}

// --- NewAuthService ---

func TestNewAuthService_ShouldReturnInstance(t *testing.T) {
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)

	svc := NewAuthService(configRepo, userRepo)

	assert.NotNil(t, svc)
}

// --- buildOAuthConfig ---

func TestAuthService_BuildOAuthConfig_WhenValidConfig_ShouldReturnConfig(t *testing.T) {
	saveAndRestoreGlobals(t)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo).(*authService)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "9999"
	configRepo.On("GetConfig").Return(cfg, nil)

	oauthCfg, err := svc.buildOAuthConfig()

	require.NoError(t, err)
	assert.Equal(t, "test-client-id", oauthCfg.ClientID)
	assert.Equal(t, "test-client-secret", oauthCfg.ClientSecret)
	assert.Equal(t, "http://localhost:9999/callback", oauthCfg.RedirectURL)
	assert.Len(t, oauthCfg.Scopes, 3)
	configRepo.AssertExpectations(t)
}

func TestAuthService_BuildOAuthConfig_WhenEmptyPort_ShouldDefaultTo8881(t *testing.T) {
	saveAndRestoreGlobals(t)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo).(*authService)

	cfg := defaultTestConfig()
	cfg.CallbackPort = ""
	configRepo.On("GetConfig").Return(cfg, nil)

	oauthCfg, err := svc.buildOAuthConfig()

	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8881/callback", oauthCfg.RedirectURL)
}

func TestAuthService_BuildOAuthConfig_WhenConfigError_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo).(*authService)

	configRepo.On("GetConfig").Return(nil, errors.New("config not found"))

	oauthCfg, err := svc.buildOAuthConfig()

	assert.Nil(t, oauthCfg)
	assert.ErrorContains(t, err, "reading config")
}

// --- GetValidToken ---

func TestAuthService_GetValidToken_WhenTokenNotExpired_ShouldReturnStoredToken(t *testing.T) {
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	token := &oauth2.Token{
		AccessToken: "valid-access",
		Expiry:      time.Now().Add(1 * time.Hour),
	}
	configRepo.On("GetGoogleToken").Return(token, nil)

	got, err := svc.GetValidToken()

	require.NoError(t, err)
	assert.Equal(t, "valid-access", got.AccessToken)
	configRepo.AssertExpectations(t)
}

func TestAuthService_GetValidToken_WhenTokenReadError_ShouldReturnError(t *testing.T) {
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	configRepo.On("GetGoogleToken").Return(nil, errors.New("no token"))

	got, err := svc.GetValidToken()

	assert.Nil(t, got)
	assert.ErrorContains(t, err, "reading stored token")
}

func TestAuthService_GetValidToken_WhenTokenExpiredAndConfigError_ShouldReturnError(t *testing.T) {
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	expiredToken := &oauth2.Token{
		AccessToken: "expired",
		Expiry:      time.Now().Add(-1 * time.Hour),
	}
	configRepo.On("GetGoogleToken").Return(expiredToken, nil)
	configRepo.On("GetConfig").Return(nil, errors.New("config error"))

	got, err := svc.GetValidToken()

	assert.Nil(t, got)
	assert.ErrorContains(t, err, "reading config")
}

func TestAuthService_GetValidToken_WhenTokenExpiredAndRefreshSucceeds_ShouldReturnNewToken(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeTokenServer(t, "refreshed-access")
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	expiredToken := &oauth2.Token{
		AccessToken:  "expired",
		RefreshToken: "old-refresh",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
	configRepo.On("GetGoogleToken").Return(expiredToken, nil)
	configRepo.On("GetConfig").Return(defaultTestConfig(), nil)
	configRepo.On("SetGoogleCredentials", mock.AnythingOfType("oauth2.Token")).Return(nil)

	got, err := svc.GetValidToken()

	require.NoError(t, err)
	assert.Equal(t, "refreshed-access", got.AccessToken)
	configRepo.AssertExpectations(t)
}

func TestAuthService_GetValidToken_WhenTokenExpiredAndRefreshFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeErrorTokenServer(t)
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	expiredToken := &oauth2.Token{
		AccessToken:  "expired",
		RefreshToken: "bad-refresh",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
	configRepo.On("GetGoogleToken").Return(expiredToken, nil)
	configRepo.On("GetConfig").Return(defaultTestConfig(), nil)

	got, err := svc.GetValidToken()

	assert.Nil(t, got)
	assert.ErrorContains(t, err, "refreshing token")
}

func TestAuthService_GetValidToken_WhenSaveRefreshedTokenFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeTokenServer(t, "refreshed")
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	expiredToken := &oauth2.Token{
		AccessToken:  "expired",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(-1 * time.Hour),
	}
	configRepo.On("GetGoogleToken").Return(expiredToken, nil)
	configRepo.On("GetConfig").Return(defaultTestConfig(), nil)
	configRepo.On("SetGoogleCredentials", mock.AnythingOfType("oauth2.Token")).Return(errors.New("disk full"))

	got, err := svc.GetValidToken()

	assert.Nil(t, got)
	assert.ErrorContains(t, err, "saving refreshed token")
}

// --- Login ---

func TestAuthService_Login_WhenBuildOAuthConfigFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	configRepo.On("GetConfig").Return(nil, errors.New("no config"))

	user, err := svc.Login()

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "reading config")
}

func TestAuthService_Login_WhenBrowserOpenFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)
	browserOpenFunc = func(_ string) error {
		return errors.New("no browser available")
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18881"
	configRepo.On("GetConfig").Return(cfg, nil)

	user, err := svc.Login()

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "opening browser")
}

func TestAuthService_Login_WhenCallbackHasNoCode_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18882"
	configRepo.On("GetConfig").Return(cfg, nil)

	browserOpenFunc = func(_ string) error { return nil }

	errCh := make(chan error, 1)
	go func() {
		_, err := svc.Login()
		errCh <- err
	}()

	simulateCallback(t, "18882", "error=access_denied", http.StatusBadRequest)

	err := <-errCh
	assert.ErrorContains(t, err, "authorization denied")
}

func TestAuthService_Login_WhenFullFlowSucceeds_ShouldReturnUser(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeTokenServer(t, "login-access-token")
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18883"
	configRepo.On("GetConfig").Return(cfg, nil)
	configRepo.On("SetGoogleCredentials", mock.AnythingOfType("oauth2.Token")).Return(nil)
	userRepo.On("GetUserInfo", "login-access-token").Return(
		&domain.User{DisplayName: "Test User", Email: "test@test.com"}, nil,
	)

	browserOpenFunc = func(_ string) error { return nil }

	resultCh := make(chan *domain.User, 1)
	errCh := make(chan error, 1)
	go func() {
		user, err := svc.Login()
		resultCh <- user
		errCh <- err
	}()

	simulateCallback(t, "18883", "code=test-auth-code", http.StatusOK)

	user := <-resultCh
	err := <-errCh

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Test User", user.DisplayName)
	assert.Equal(t, "test@test.com", user.Email)
	configRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestAuthService_Login_WhenExchangeFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeErrorTokenServer(t)
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18884"
	configRepo.On("GetConfig").Return(cfg, nil)

	browserOpenFunc = func(_ string) error { return nil }

	errCh := make(chan error, 1)
	go func() {
		_, err := svc.Login()
		errCh <- err
	}()

	simulateCallback(t, "18884", "code=bad-code", http.StatusOK)

	err := <-errCh
	assert.ErrorContains(t, err, "exchanging code for token")
}

func TestAuthService_Login_WhenSaveTokenFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeTokenServer(t, "token")
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18885"
	configRepo.On("GetConfig").Return(cfg, nil)
	configRepo.On("SetGoogleCredentials", mock.AnythingOfType("oauth2.Token")).Return(errors.New("disk full"))

	browserOpenFunc = func(_ string) error { return nil }

	errCh := make(chan error, 1)
	go func() {
		_, err := svc.Login()
		errCh <- err
	}()

	simulateCallback(t, "18885", "code=test-code", http.StatusOK)

	err := <-errCh
	assert.ErrorContains(t, err, "saving token")
}

func TestAuthService_Login_WhenFetchUserInfoFails_ShouldReturnError(t *testing.T) {
	saveAndRestoreGlobals(t)

	tokenServer := fakeTokenServer(t, "token-for-userinfo")
	defer tokenServer.Close()

	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "http://localhost/auth",
		TokenURL: tokenServer.URL,
	}

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = "18886"
	configRepo.On("GetConfig").Return(cfg, nil)
	configRepo.On("SetGoogleCredentials", mock.AnythingOfType("oauth2.Token")).Return(nil)
	userRepo.On("GetUserInfo", "token-for-userinfo").Return(nil, errors.New("API down"))

	browserOpenFunc = func(_ string) error { return nil }

	errCh := make(chan error, 1)
	go func() {
		_, err := svc.Login()
		errCh <- err
	}()

	simulateCallback(t, "18886", "code=test-code", http.StatusOK)

	err := <-errCh
	assert.ErrorContains(t, err, "fetching user info")
}

func TestAuthService_Login_WhenEmptyPort_ShouldDefaultTo8881(t *testing.T) {
	saveAndRestoreGlobals(t)

	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewAuthService(configRepo, userRepo)

	cfg := defaultTestConfig()
	cfg.CallbackPort = ""
	configRepo.On("GetConfig").Return(cfg, nil)

	browserOpenFunc = func(_ string) error {
		return errors.New("stop")
	}

	user, err := svc.Login()

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "opening browser")
}
