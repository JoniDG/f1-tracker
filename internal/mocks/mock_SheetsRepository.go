package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockSheetsRepository struct {
	mock.Mock
}

func (m *MockSheetsRepository) GetSheetValues(accessToken, spreadsheetID, sheetName string) error {
	args := m.Called(accessToken, spreadsheetID, sheetName)
	return args.Error(0)
}
