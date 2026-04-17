package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	HasValidConfig() bool
}

type configRepository struct {
	viper        *viper.Viper
	configLoaded bool
}

func NewConfigRepository() (ConfigRepository, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("getting user config dir: %w", err)
	}

	configPath := filepath.Join(userConfigDir, defines.ConfigPath)
	if err = os.MkdirAll(configPath, 0o750); err != nil {
		return nil, fmt.Errorf("creating config dir: %w", err)
	}

	v := viper.New()
	v.SetConfigType("json")
	v.AddConfigPath(configPath)
	v.SetConfigName(defines.ConfigFilename)

	cr := &configRepository{
		viper:        v,
		configLoaded: true,
	}

	if err = v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
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
	r.viper.Set("token", token)
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
	if err := r.viper.UnmarshalKey("config", &config); err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	return &config, nil
}

func (r *configRepository) HasValidConfig() bool {
	cfg, err := r.GetConfig()
	if err != nil {
		return false
	}

	return cfg.GoogleClientID != "" &&
		cfg.GoogleClientSecret != "" &&
		strings.HasSuffix(cfg.GoogleClientID, ".apps.googleusercontent.com")
}
