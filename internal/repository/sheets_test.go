package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSheetsRepository_ShouldReturnInstance(t *testing.T) {
	repo := NewSheetsRepository()

	assert.NotNil(t, repo)
}
