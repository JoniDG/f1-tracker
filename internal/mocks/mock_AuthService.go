package mocks

import (
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login() (*domain.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GetValidToken() (*oauth2.Token, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth2.Token), args.Error(1)
}

func (m *MockAuthService) HasValidConfig() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthService) HasStoredToken() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAuthService) GetConfig() (*domain.Config, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Config), args.Error(1)
}

func (m *MockAuthService) SetConfig(c domain.Config) error {
	args := m.Called(c)
	return args.Error(0)
}
