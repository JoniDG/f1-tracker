package repository

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/go-resty/resty/v2"
)

type UserRepository interface {
	// GetUserInfo consulta la API de Google para obtener nombre y email del usuario autenticado.
	GetUserInfo(accessToken string) (*domain.User, error)
}

type userRepository struct {
	baseURL string
}

func NewUserRepository() UserRepository {
	return &userRepository{
		baseURL: "https://www.googleapis.com",
	}
}

func (r *userRepository) GetUserInfo(accessToken string) (*domain.User, error) {
	// resty es un HTTP client que simplifica las llamadas REST.
	// SetAuthToken agrega el header "Authorization: Bearer {token}" automaticamente.
	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		Get(r.baseURL + "/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("calling userinfo API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("userinfo API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	// Parsea la respuesta JSON de Google al struct domain.User
	var userInfo domain.User
	if err := json.Unmarshal(resp.Body(), &userInfo); err != nil {
		return nil, fmt.Errorf("parsing userinfo response: %w", err)
	}

	return &userInfo, nil
}
