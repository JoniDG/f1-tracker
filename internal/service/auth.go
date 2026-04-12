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

type AuthService interface {
	// Login ejecuta el flujo completo de OAuth2+PKCE con Google.
	// Es bloqueante: espera hasta que el usuario autorice en el browser.
	// La capa UI debe llamarlo en una goroutine para no bloquear la interfaz.
	Login() (*domain.User, error)
	// GetValidToken devuelve un token valido. Si el access token expiro,
	// usa el refresh token para obtener uno nuevo y lo persiste en config.
	GetValidToken() (*oauth2.Token, error)
}

type authService struct {
	configRepo repository.ConfigRepository
	userRepo   repository.UserRepository
}

func NewAuthService(configRepo repository.ConfigRepository, userRepo repository.UserRepository) AuthService {
	return &authService{configRepo: configRepo, userRepo: userRepo}
}

// buildOAuthConfig construye el oauth2.Config a partir de la config guardada.
// Se reutiliza en Login (para el flujo completo) y en GetValidToken (para refrescar).
func (s *authService) buildOAuthConfig() (*oauth2.Config, error) {
	cfg, err := s.configRepo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	port := cfg.CallbackPort
	if port == "" {
		port = "8881"
	}

	// oauth2.Config define los parametros de la autenticacion OAuth2:
	// - ClientID/Secret: credenciales de la app en Google Cloud Console
	// - Endpoint: URLs de Google para autorizar y obtener tokens
	// - RedirectURL: a donde Google redirige despues de autorizar (nuestro servidor local)
	// - Scopes: permisos que pedimos (spreadsheets para leer/escribir hojas, userinfo para nombre/email)
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		RedirectURL: fmt.Sprintf("http://localhost:%s/callback", port),
		Scopes: []string{
			"https://www.googleapis.com/auth/spreadsheets",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}, nil
}

func (s *authService) GetValidToken() (*oauth2.Token, error) {
	// Lee el token guardado en config
	token, err := s.configRepo.GetGoogleToken()
	if err != nil {
		return nil, fmt.Errorf("reading stored token: %w", err)
	}

	// Si el token no expiro, lo devuelve tal cual
	if token.Expiry.After(time.Now()) {
		return token, nil
	}

	// El access token expiro, necesitamos refrescarlo usando el refresh token.
	// Para eso necesitamos el oauth2.Config con ClientID/Secret y el TokenURL.
	oauthConfig, err := s.buildOAuthConfig()
	if err != nil {
		return nil, err
	}

	// TokenSource crea un source que automaticamente usa el refresh token
	// para pedir un nuevo access token al TokenURL de Google.
	// .Token() ejecuta el refresh y devuelve el token nuevo.
	newToken, err := oauthConfig.TokenSource(context.Background(), token).Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}

	// Persiste el token nuevo (con el nuevo access token y expiry) en config
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

	// Extraemos el puerto de la config para el servidor de callback
	cfg, _ := s.configRepo.GetConfig()
	port := cfg.CallbackPort
	if port == "" {
		port = "8881"
	}

	// PKCE (Proof Key for Code Exchange) agrega seguridad al flujo OAuth2:
	// - GenerateVerifier crea un string aleatorio (el "verifier")
	// - S256ChallengeOption genera un hash SHA256 del verifier (el "challenge")
	// - El challenge se envia con la URL de auth, el verifier se envia al intercambiar el code
	// - Google verifica que el que pide el token sea el mismo que inicio el flujo
	verifier := oauth2.GenerateVerifier()

	// AuthCodeURL construye la URL de autorizacion de Google con todos los parametros:
	// - "state": valor anti-CSRF (para MVP usamos un valor fijo)
	// - AccessTypeOffline: pide un refresh_token ademas del access_token
	// - prompt=consent: fuerza a Google a mostrar la pantalla de permisos siempre
	// - S256ChallengeOption: incluye el challenge PKCE en la URL
	authURL := oauthConfig.AuthCodeURL(
		"state",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"),
		oauth2.S256ChallengeOption(verifier),
	)

	// Canales con buffer 1 para comunicar la goroutine del servidor HTTP con el flujo principal.
	// callbackCode recibe el authorization code que Google envia al callback.
	// callbackErr recibe errores (usuario denego acceso, servidor no pudo arrancar, etc).
	callbackCode := make(chan string, 1)
	callbackErr := make(chan error, 1)

	// Gin en modo release para no imprimir logs de debug en la consola
	gin.SetMode(gin.ReleaseMode)
	// gin.New() crea un router sin middlewares (no necesitamos logger ni recovery para un server temporal)
	router := gin.New()

	// Endpoint /callback: Google redirige aca despues de que el usuario autoriza.
	// El query param "code" contiene el authorization code que necesitamos para obtener el token.
	router.GET("/callback", func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			// Si no hay code, el usuario denego el acceso o hubo un error
			errMsg := c.Query("error")
			c.Header("Content-Type", "text/html")
			c.String(http.StatusBadRequest, "<h1>Error</h1><p>%s</p>", errMsg)
			callbackErr <- fmt.Errorf("authorization denied: %s", errMsg)
			return
		}
		// Responde al browser con un mensaje para que el usuario sepa que puede cerrar la pestana
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, "<h1>Listo!</h1><p>Podes cerrar esta pestana y volver a la app.</p>")
		// Envia el code al canal para que el flujo principal lo reciba
		callbackCode <- code
	})

	// Wrapeamos el router de Gin en un http.Server para poder llamar Shutdown() despues.
	// Si usaramos router.Run() directamente, no podriamos detener el servidor limpiamente.
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Lanza el servidor HTTP en una goroutine para que no bloquee el flujo principal.
	// ListenAndServe es bloqueante: escucha conexiones hasta que se llame Shutdown().
	// Si el puerto esta ocupado, envia el error por el canal.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			callbackErr <- fmt.Errorf("callback server error: %w", err)
		}
	}()

	// Abre el browser del usuario con la URL de autorizacion de Google
	if err := browser.OpenURL(authURL); err != nil {
		_ = server.Shutdown(context.Background())
		return nil, fmt.Errorf("opening browser: %w", err)
	}

	// select espera el primer valor que llegue de cualquiera de los dos canales:
	// - Si llega el code: el usuario autorizo, continuamos
	// - Si llega un error: algo fallo, cerramos el servidor y retornamos el error
	var code string
	select {
	case code = <-callbackCode:
	case err := <-callbackErr:
		_ = server.Shutdown(context.Background())
		return nil, err
	}

	// Ya tenemos el code, apagamos el servidor HTTP porque no lo necesitamos mas.
	// Shutdown cierra el servidor esperando que las conexiones activas terminen.
	_ = server.Shutdown(context.Background())

	// Exchange intercambia el authorization code por un access_token (y refresh_token).
	// VerifierOption envia el PKCE verifier original para que Google valide el flujo.
	ctx := context.Background()
	token, err := oauthConfig.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("exchanging code for token: %w", err)
	}

	// Persiste el token completo (access, refresh, expiry) en el archivo de config local
	if err := s.configRepo.SetGoogleCredentials(*token); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}

	// Consulta la API de Google para obtener el nombre y email del usuario autenticado
	user, err := s.userRepo.GetUserInfo(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}

	return user, nil
}
