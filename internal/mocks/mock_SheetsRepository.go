package mocks

import (
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

type MockSheetsRepository struct {
	mock.Mock
}

func (m *MockSheetsRepository) GetSheetValues(accessToken, spreadsheetID, sheetName string) error {
	args := m.Called(accessToken, spreadsheetID, sheetName)
	return args.Error(0)
}

func (m *MockSheetsRepository) GetSpreadsheetData(accessToken, spreadsheetID string) (*domain.SpreadsheetData, error) {
	args := m.Called(accessToken, spreadsheetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SpreadsheetData), args.Error(1)
}

func (m *MockSheetsRepository) AddSheet(accessToken, spreadsheetID, sheetName string) error {
	args := m.Called(accessToken, spreadsheetID, sheetName)
	return args.Error(0)
}

func (m *MockSheetsRepository) UpdateSheetValues(accessToken, spreadsheetID, rangeValues string, values [][]string) error {
	args := m.Called(accessToken, spreadsheetID, rangeValues, values)
	return args.Error(0)
}

func (m *MockSheetsRepository) CreateSpreadsheet(accessToken, title string) (string, error) {
	args := m.Called(accessToken, title)
	return args.String(0), args.Error(1)
}
