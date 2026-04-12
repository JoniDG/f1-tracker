package service

import (
	"errors"
	"testing"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewTrackerService_ShouldReturnInstance(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)

	svc := NewTrackerService(authSvc, configRepo, userRepo)

	assert.NotNil(t, svc)
}

func TestTrackerService_GetCurrentUser_WhenValidToken_ShouldReturnUser(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	expectedUser := &domain.User{DisplayName: "Juan", Email: "juan@test.com"}

	authSvc.On("GetValidToken").Return(token, nil)
	userRepo.On("GetUserInfo", "valid-token").Return(expectedUser, nil)

	user, err := svc.GetCurrentUser()

	require.NoError(t, err)
	assert.Equal(t, "Juan", user.DisplayName)
	assert.Equal(t, "juan@test.com", user.Email)
	authSvc.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestTrackerService_GetCurrentUser_WhenTokenError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo)

	authSvc.On("GetValidToken").Return(nil, errors.New("token expired"))

	user, err := svc.GetCurrentUser()

	assert.Nil(t, user)
	assert.EqualError(t, err, "token expired")
	authSvc.AssertExpectations(t)
}

func TestTrackerService_GetCurrentUser_WhenUserInfoError_ShouldReturnError(t *testing.T) {
	authSvc := new(mocks.MockAuthService)
	configRepo := new(mocks.MockConfigRepository)
	userRepo := new(mocks.MockUserRepository)
	svc := NewTrackerService(authSvc, configRepo, userRepo)

	token := &oauth2.Token{AccessToken: "valid-token"}
	authSvc.On("GetValidToken").Return(token, nil)
	userRepo.On("GetUserInfo", "valid-token").Return(nil, errors.New("API error"))

	user, err := svc.GetCurrentUser()

	assert.Nil(t, user)
	assert.EqualError(t, err, "API error")
	authSvc.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}
