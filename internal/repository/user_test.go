package repository

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserRepository_ShouldReturnInstance(t *testing.T) {
	repo := NewUserRepository()

	assert.NotNil(t, repo)
}

func TestUserRepository_GetUserInfo_WhenSuccess_ShouldReturnUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "/oauth2/v2/userinfo", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"Juan Perez","email":"juan@test.com"}`))
	}))
	defer server.Close()

	repo := newUserRepository(server.URL)

	user, err := repo.GetUserInfo("test-token")

	require.NoError(t, err)
	assert.Equal(t, "Juan Perez", user.DisplayName)
	assert.Equal(t, "juan@test.com", user.Email)
}

func TestUserRepository_GetUserInfo_WhenNon200Status_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_token"}`))
	}))
	defer server.Close()

	repo := newUserRepository(server.URL)

	user, err := repo.GetUserInfo("bad-token")

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "userinfo API returned status 401")
}

func TestUserRepository_GetUserInfo_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := newUserRepository(server.URL)

	user, err := repo.GetUserInfo("test-token")

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "parsing userinfo response")
}

func TestUserRepository_GetUserInfo_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := newUserRepository("http://localhost:1")

	user, err := repo.GetUserInfo("test-token")

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "calling userinfo API")
}
