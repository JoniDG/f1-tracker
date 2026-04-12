package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JoniDG/f1-tracker/internal/defines"
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type ConfigRepository interface {
	GetGoogleToken() (*oauth2.Token, error)
	SetGoogleCredentials(token oauth2.Token) error
	SetConfig(c domain.Config) error
	GetConfig() (*domain.Config, error)
	IsLoaded() bool
}

type configRepository struct {
	viper        *viper.Viper
	configLoaded bool
}

func NewConfigRepository() (ConfigRepository, error) {
	// Obtiene el directorio de configuración del SO (~/.config en Linux/Mac, AppData en Windows)
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting user config dir: %w", err)
	}

	// Arma el path completo: ~/.config/f1-tracker/
	configPath := filepath.Join(userConfigDir, defines.ConfigPath)
	// Crea el directorio (y cualquier padre faltante) con permisos rwxr-xr-x
	if err = os.MkdirAll(configPath, 0o750); err != nil {
		return nil, fmt.Errorf("creating config dir: %w", err)
	}

	// Crea una nueva instancia de viper (librería para manejar archivos de configuración)
	v := viper.New()
	// Indica que el archivo de config es JSON
	v.SetConfigType("json")
	// Le dice a viper en qué directorio buscar el archivo
	v.AddConfigPath(configPath)
	// Nombre del archivo sin extensión (viper le agrega .json por el SetConfigType)
	v.SetConfigName(defines.ConfigFilename)

	cr := &configRepository{
		viper:        v,
		configLoaded: true,
	}

	// Intenta leer el archivo de config existente y cargarlo en memoria
	if err = v.ReadInConfig(); err != nil {
		// Si el error es que no existe el archivo, lo creamos; cualquier otro error es fatal
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
		// SafeWriteConfig crea el archivo solo si no existe (no sobreescribe uno existente)
		if err = v.SafeWriteConfig(); err != nil {
			return nil, fmt.Errorf("creating initial config file: %w", err)
		}
		cr.configLoaded = false
	}

	return cr, nil
}

func (r *configRepository) GetGoogleToken() (*oauth2.Token, error) {
	if !r.configLoaded {
		return nil, errors.New("no config found, run 'f1-tracker config' first")
	}

	// No usamos UnmarshalKey aca porque mapstructure no sabe convertir
	// un string ISO ("2026-04-12T...") a time.Time (que es el tipo de Expiry).
	// En cambio, serializamos a JSON y deserializamos con encoding/json que si lo soporta.
	tokenMap := r.viper.Get("token")
	jsonBytes, err := json.Marshal(tokenMap)
	if err != nil {
		return nil, fmt.Errorf("reading google token: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(jsonBytes, &token); err != nil {
		return nil, fmt.Errorf("parsing google token: %w", err)
	}

	return &token, nil
}

func (r *configRepository) SetGoogleCredentials(token oauth2.Token) error {
	if !r.configLoaded {
		return errors.New("no config found, run 'f1-tracker config' first")
	}
	// Set guarda el valor en memoria (no en disco)
	r.viper.Set("token", token)
	// WriteConfig persiste todos los valores en memoria al archivo JSON en disco
	return r.viper.WriteConfig()
}

func (r *configRepository) SetConfig(c domain.Config) error {
	r.configLoaded = true
	r.viper.Set("config", c)
	return r.viper.WriteConfig()
}

func (r *configRepository) GetConfig() (*domain.Config, error) {
	if !r.configLoaded {
		return nil, errors.New("no config found, run 'f1-tracker config' first")
	}

	var config domain.Config
	// UnmarshalKey lee la key "config" del JSON y mapea los campos al struct domain.Config
	if err := r.viper.UnmarshalKey("config", &config); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return &config, nil
}

func (r *configRepository) IsLoaded() bool {
	return r.configLoaded
}
