package mocks

import (
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) GetGoogleToken() (*oauth2.Token, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *MockConfigRepository) SetGoogleCredentials(token oauth2.Token) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockConfigRepository) SetConfig(c domain.Config) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockConfigRepository) GetConfig() (*domain.Config, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Config), args.Error(1)
}

func (m *MockConfigRepository) IsLoaded() bool {
	args := m.Called()
	return args.Bool(0)
}
