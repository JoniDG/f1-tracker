package repository

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/go-resty/resty/v2"
)

type SheetsRepository interface {
	GetSheetValues(accessToken, spreadsheetID, sheetName string) error
	GetSpreadsheetData(accessToken, spreadsheetID string) (*domain.SpreadsheetData, error)
	AddSheet(accessToken, spreadsheetID, sheetName string) error
	UpdateSheetValues(accessToken, spreadsheetID, rangeValues string, values [][]string) error
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
		return fmt.Errorf("calling spreadsheet API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("spreadsheet API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var result struct {
		Range          string     `json:"range"`
		MajorDimension string     `json:"majorDimension"`
		Values         [][]string `json:"values"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return fmt.Errorf("parsing spreadsheet response: %w", err)
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
		return nil, fmt.Errorf("calling spreadsheet API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("spreadsheet API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var spreadsheetData domain.SpreadsheetData
	if err := json.Unmarshal(resp.Body(), &spreadsheetData); err != nil {
		return nil, fmt.Errorf("parsing spreadsheet response: %w", err)
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
		return fmt.Errorf("calling batchUpdate API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("batchUpdate API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	var response domain.BatchResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("parsing batchUpdate API response: %w", err)
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
		return fmt.Errorf("calling spreadsheet API: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("spreadsheet API returned status %d: %s", resp.StatusCode(), resp.String())
	}
	var response domain.UpdateSheetValuesResponse
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("parsing spreadsheet API response: %w", err)
	}
	return nil
}

func (r *sheetsRepository) CreateSpreadsheet(accessToken, spreadsheetID string) error {
	return nil
}

//printJSONForCopy(resp.Body())
//debugJSON(resp.Body())

func debugJSON(data []byte) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber() // 🔑 evita que todos los números sean float64

	var parsed interface{}
	if err := decoder.Decode(&parsed); err != nil {
		fmt.Println("❌ Error parseando JSON:", err)
		fmt.Println("Raw:", string(data))
		return
	}

	printWithTypes(parsed, 0)
}

// 🔍 Recursivo con tipos
func printWithTypes(v interface{}, indent int) {
	prefix := spaces(indent)

	switch val := v.(type) {

	case map[string]interface{}:
		fmt.Println(prefix + "{object}")
		for k, v2 := range val {
			fmt.Printf("%s  %s: ", prefix, k)
			printWithTypes(v2, indent+2)
		}

	case []interface{}:
		fmt.Println(prefix + "[array]")
		for i, item := range val {
			fmt.Printf("%s  [%d]: ", prefix, i)
			printWithTypes(item, indent+2)
		}

	case string:
		fmt.Printf("%s(string) %q\n", prefix, val)

	case json.Number:
		// Detectar si es int o float
		if _, err := val.Int64(); err == nil {
			fmt.Printf("%s(int) %s\n", prefix, val.String())
		} else {
			fmt.Printf("%s(float) %s\n", prefix, val.String())
		}

	case bool:
		fmt.Printf("%s(bool) %v\n", prefix, val)

	case nil:
		fmt.Printf("%s(null)\n", prefix)

	default:
		fmt.Printf("%s(%s) %v\n", prefix, reflect.TypeOf(val), val)
	}
}
func printJSONForCopy(data []byte) {
	var out bytes.Buffer

	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		fmt.Println("❌ No es JSON válido, imprimiendo raw:")
		fmt.Println(string(data))
		return
	}

	fmt.Println("📋 --- JSON (copy/paste ready) ---")
	fmt.Println(out.String())
	fmt.Println("📋 --- END ---")
}
func spaces(n int) string {
	return fmt.Sprintf("%*s", n, "")
}
