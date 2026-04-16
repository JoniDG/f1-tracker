package repository

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/go-resty/resty/v2"
)

type SheetsRepository interface {
	GetSheetValues(accessToken, spreadsheetID, sheetName string) error
	GetSpreadsheetData(accessToken, spreadsheetID string) (*domain.SpreadsheetData, error)
	AddSheet(accessToken, spreadsheetID, sheetName string) error
	UpdateSheetValues(accessToken, spreadsheetID, rangeValues string, values [][]string) error
	CreateSpreadsheet(accessToken, title string) (string, error)
}

type sheetsRepository struct {
	baseURL string
}

func NewSheetsRepository() SheetsRepository {
	return &sheetsRepository{
		baseURL: "https://sheets.googleapis.com/v4/spreadsheets",
	}
}

func (r *sheetsRepository) GetSheetValues(accessToken, spreadsheetID, sheetName string) error {
	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		SetPathParam("spreadsheetID", spreadsheetID).
		SetPathParam("sheetName", sheetName).
		Get(r.baseURL + "/{spreadsheetID}/values/{sheetName}")
	if err != nil {
		return fmt.Errorf("calling get sheet values API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("get sheet values API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var result domain.GetSheetValuesResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("parsing get sheet values response: %w", err)
	}

	for _, row := range result.Values {
		fmt.Println(row)
	}

	return nil
}
func (r *sheetsRepository) GetSpreadsheetData(accessToken, spreadsheetID string) (*domain.SpreadsheetData, error) {
	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		SetPathParam("spreadsheetID", spreadsheetID).
		SetQueryParam("fields", "spreadsheetId,properties,sheets.properties").
		Get(r.baseURL + "/{spreadsheetID}")
	if err != nil {
		return nil, fmt.Errorf("calling get spreadsheet data API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("get spreadsheet data API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var spreadsheetData domain.SpreadsheetData
	if err := json.Unmarshal(resp.Body(), &spreadsheetData); err != nil {
		return nil, fmt.Errorf("parsing get spreadsheet data response: %w", err)
	}

	return &spreadsheetData, nil
}
func (r *sheetsRepository) AddSheet(accessToken, spreadsheetID, sheetName string) error {
	body := domain.BatchUpdateRequest{
		Requests: []domain.BatchRequest{
			{
				AddSheet: &domain.AddSheetRequest{
					Properties: domain.AddSheetRequestProperties{Title: sheetName},
				},
			},
		},
	}

	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		SetPathParam("spreadsheetID", spreadsheetID).
		SetBody(body).
		Post(r.baseURL + "/{spreadsheetID}:batchUpdate")
	if err != nil {
		return fmt.Errorf("calling add sheet API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("add sheet API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var response domain.BatchResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("parsing add sheet response: %w", err)
	}
	return nil
}
func (r *sheetsRepository) UpdateSheetValues(accessToken, spreadsheetID, rangeValues string, values [][]string) error {
	body := domain.UpdateSheetValuesRequest{
		Range:          rangeValues,
		MajorDimension: "ROWS",
		Values:         values,
	}
	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		SetPathParam("spreadsheetID", spreadsheetID).
		SetPathParam("range", rangeValues).
		SetQueryParam("valueInputOption", "USER_ENTERED").
		SetBody(body).
		Put(r.baseURL + "/{spreadsheetID}/values/{range}")
	if err != nil {
		return fmt.Errorf("calling update sheet values API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("update sheet values API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var response domain.UpdateSheetValuesResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("parsing update sheet values response: %w", err)
	}
	return nil
}
func (r *sheetsRepository) CreateSpreadsheet(accessToken, title string) (string, error) {
	body := domain.CreateSpreadsheetRequest{
		Properties: domain.CreateSpreadsheetRequestProperties{Title: title},
	}

	resp, err := resty.New().R().
		SetAuthToken(accessToken).
		SetBody(body).
		Post(r.baseURL)
	if err != nil {
		return "", fmt.Errorf("calling create spreadsheet API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("create spreadsheet API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var result domain.CreateSpreadsheetResponse
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return "", fmt.Errorf("parsing create spreadsheet response: %w", err)
	}

	return result.SpreadsheetId, nil
}
