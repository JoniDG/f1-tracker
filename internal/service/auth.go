package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

var (
	browserOpenFunc = browser.OpenURL
	googleEndpoint  = oauth2.Endpoint{ // #nosec G101
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
	}
)

type AuthService interface {
	Login() (*domain.User, error)
	GetValidToken() (*oauth2.Token, error)
	HasValidConfig() bool
	HasStoredToken() bool
	GetConfig() (*domain.Config, error)
	SetConfig(c domain.Config) error
}

type authService struct {
	configRepo repository.ConfigRepository
	userRepo   repository.UserRepository
}

func NewAuthService(configRepo repository.ConfigRepository, userRepo repository.UserRepository) AuthService {
	return &authService{configRepo: configRepo, userRepo: userRepo}
}

func (s *authService) HasValidConfig() bool {
	return s.configRepo.HasValidConfig()
}

func (s *authService) HasStoredToken() bool {
	_, err := s.configRepo.GetGoogleToken()
	return err == nil
}

func (s *authService) GetConfig() (*domain.Config, error) {
	return s.configRepo.GetConfig()
}

func (s *authService) SetConfig(c domain.Config) error {
	return s.configRepo.SetConfig(c)
}

func (s *authService) buildOAuthConfig() (*oauth2.Config, error) {
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	port := cfg.CallbackPort
	if port == "" {
		port = "8081"
	}

	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint:     googleEndpoint,
		RedirectURL:  fmt.Sprintf("http://localhost:%s/callback", port),
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}, nil
}

func (s *authService) GetValidToken() (*oauth2.Token, error) {
	token, err := s.configRepo.GetGoogleToken()
	if err != nil {
		return nil, fmt.Errorf("reading stored token: %w", err)
	}

	if token.Expiry.After(time.Now()) {
		return token, nil
	}

	oauthConfig, err := s.buildOAuthConfig()
	if err != nil {
		return nil, err
	}

	newToken, err := oauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	if err := s.configRepo.SetGoogleCredentials(*newToken); err != nil {
		return nil, fmt.Errorf("saving refreshed token: %w", err)
	}

	return newToken, nil
}

func (s *authService) Login() (*domain.User, error) {
	oauthConfig, err := s.buildOAuthConfig()
	if err != nil {
		return nil, err
	}

	cfg, _ := s.configRepo.GetConfig()
	port := cfg.CallbackPort
	if port == "" {
		port = "8081"
	}

	verifier := oauth2.GenerateVerifier()

	authURL := oauthConfig.AuthCodeURL(
		"state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.S256ChallengeOption(verifier),
	)

	callbackCode := make(chan string, 1)
	callbackErr := make(chan error, 1)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			errMsg := c.Query("error")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, "<h1>Error</h1><p>%s</p>", errMsg)
			callbackErr <- fmt.Errorf("authorization denied: %s", errMsg)
			return
		}
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, "<h1>Listo!</h1><p>Podes cerrar esta pestana y volver a la app.</p>")
		callbackCode <- code
	})

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			callbackErr <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	if err := browserOpenFunc(authURL); err != nil {
		_ = server.Shutdown(context.Background())
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	var code string
	select {
	case code = <-callbackCode:
	case err := <-callbackErr:
		_ = server.Shutdown(context.Background())
		return nil, err
	}

	_ = server.Shutdown(context.Background())

	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("exchanging code for token: %w", err)
	}

	if err := s.configRepo.SetGoogleCredentials(*token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	user, err := s.userRepo.GetUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}

	return user, nil
}
