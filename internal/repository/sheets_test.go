package repository

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSheetsRepository_ShouldReturnInstance(t *testing.T) {
	repo := NewSheetsRepository()

	assert.NotNil(t, repo)
}

func TestSheetsRepository_GetSheetValues_WhenSuccess_ShouldReturnValues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Contains(t, r.URL.Path, "/values/")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.GetSheetValuesResponse{
			Range:          "Sheet1!A1:J2",
			MajorDimension: "ROWS",
			Values:         [][]string{{"Circuito", "Mejor Vuelta"}, {"Bahrain", "1:23.456"}},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	values, err := repo.GetSheetValues("test-token", "sheet-123", "Alice")

	require.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "Bahrain", values[1][0])
}

func TestSheetsRepository_GetSheetValues_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	values, err := repo.GetSheetValues("test-token", "sheet-123", "Alice")

	assert.Nil(t, values)
	assert.ErrorContains(t, err, "get sheet values API returned status 404")
}

func TestSheetsRepository_GetSheetValues_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	values, err := repo.GetSheetValues("test-token", "sheet-123", "Alice")

	assert.Nil(t, values)
	assert.ErrorContains(t, err, "parsing get sheet values response")
}

func TestSheetsRepository_GetSheetValues_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	values, err := repo.GetSheetValues("test-token", "sheet-123", "Alice")

	assert.Nil(t, values)
	assert.ErrorContains(t, err, "calling get sheet values API")
}

func TestSheetsRepository_GetSpreadsheetData_WhenSuccess_ShouldReturnData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.SpreadsheetData{
			SpreadsheetId: "sheet-123",
			Sheets: []domain.SheetData{
				{Properties: domain.SheetDataProperties{Title: "Alice"}},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	data, err := repo.GetSpreadsheetData("test-token", "sheet-123")

	require.NoError(t, err)
	assert.Equal(t, "sheet-123", data.SpreadsheetId)
	assert.Len(t, data.Sheets, 1)
	assert.Equal(t, "Alice", data.Sheets[0].Properties.Title)
}

func TestSheetsRepository_GetSpreadsheetData_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	data, err := repo.GetSpreadsheetData("bad-token", "sheet-123")

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "get spreadsheet data API returned status 401")
}

func TestSheetsRepository_GetSpreadsheetData_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	data, err := repo.GetSpreadsheetData("test-token", "sheet-123")

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "parsing get spreadsheet data response")
}

func TestSheetsRepository_GetSpreadsheetData_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	data, err := repo.GetSpreadsheetData("test-token", "sheet-123")

	assert.Nil(t, data)
	assert.ErrorContains(t, err, "calling get spreadsheet data API")
}

func TestSheetsRepository_AddSheet_WhenSuccess_ShouldReturnSheetID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.BatchResponse{
			SpreadsheetId: "sheet-123",
			Replies: []domain.BatchReplies{
				{AddSheet: &domain.AddSheetResponse{Properties: domain.AddSheetResponseProperties{SheetId: 42, Title: "NuevaHoja"}}},
			},
		}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	sheetID, err := repo.AddSheet("test-token", "sheet-123", "NuevaHoja")

	require.NoError(t, err)
	assert.Equal(t, 42, sheetID)
}

func TestSheetsRepository_AddSheet_WhenMissingReply_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.BatchResponse{SpreadsheetId: "sheet-123"}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	_, err := repo.AddSheet("test-token", "sheet-123", "NuevaHoja")

	assert.ErrorContains(t, err, "add sheet response missing sheetId")
}

func TestSheetsRepository_AddSheet_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	_, err := repo.AddSheet("test-token", "sheet-123", "NuevaHoja")

	assert.ErrorContains(t, err, "add sheet API returned status 400")
}

func TestSheetsRepository_AddSheet_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	_, err := repo.AddSheet("test-token", "sheet-123", "NuevaHoja")

	assert.ErrorContains(t, err, "parsing add sheet response")
}

func TestSheetsRepository_AddSheet_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	_, err := repo.AddSheet("test-token", "sheet-123", "NuevaHoja")

	assert.ErrorContains(t, err, "calling add sheet API")
}

func TestSheetsRepository_AddProtectedRange_WhenSuccess_ShouldReturnNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPost, r.Method)
		var body domain.BatchUpdateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Len(t, body.Requests, 1)
		require.NotNil(t, body.Requests[0].AddProtectedRange)
		pr := body.Requests[0].AddProtectedRange.ProtectedRange
		assert.Equal(t, 42, pr.Range.SheetId)
		assert.Equal(t, "f1-tracker: Alice", pr.Description)
		assert.False(t, pr.WarningOnly)
		assert.Equal(t, []string{"alice@test.com"}, pr.Editors.Users)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	err := repo.AddProtectedRange("test-token", "sheet-123", 42, "alice@test.com", "f1-tracker: Alice")

	require.NoError(t, err)
}

func TestSheetsRepository_AddProtectedRange_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	err := repo.AddProtectedRange("test-token", "sheet-123", 42, "alice@test.com", "desc")

	assert.ErrorContains(t, err, "add protected range API returned status 403")
}

func TestSheetsRepository_AddProtectedRange_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	err := repo.AddProtectedRange("test-token", "sheet-123", 42, "alice@test.com", "desc")

	assert.ErrorContains(t, err, "calling add protected range API")
}

func TestSheetsRepository_UpdateSheetValues_WhenSuccess_ShouldReturnNil(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.UpdateSheetValuesResponse{SpreadsheetId: "sheet-123", UpdatedCells: 10}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	values := [][]string{{"Bahrain", "1:23.456"}}
	err := repo.UpdateSheetValues("test-token", "sheet-123", "Alice!A2:J2", values)

	require.NoError(t, err)
}

func TestSheetsRepository_UpdateSheetValues_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	err := repo.UpdateSheetValues("test-token", "sheet-123", "Alice!A2:J2", [][]string{{"data"}})

	assert.ErrorContains(t, err, "update sheet values API returned status 403")
}

func TestSheetsRepository_UpdateSheetValues_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	err := repo.UpdateSheetValues("test-token", "sheet-123", "Alice!A2:J2", [][]string{{"data"}})

	assert.ErrorContains(t, err, "parsing update sheet values response")
}

func TestSheetsRepository_UpdateSheetValues_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	err := repo.UpdateSheetValues("test-token", "sheet-123", "Alice!A2:J2", [][]string{{"data"}})

	assert.ErrorContains(t, err, "calling update sheet values API")
}

func TestSheetsRepository_CreateSpreadsheet_WhenSuccess_ShouldReturnID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		assert.Equal(t, http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := domain.CreateSpreadsheetResponse{SpreadsheetId: "new-sheet-456"}
		require.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	id, err := repo.CreateSpreadsheet("test-token", "Mi Spreadsheet")

	require.NoError(t, err)
	assert.Equal(t, "new-sheet-456", id)
}

func TestSheetsRepository_CreateSpreadsheet_WhenNon200_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	id, err := repo.CreateSpreadsheet("test-token", "Mi Spreadsheet")

	assert.Empty(t, id)
	assert.ErrorContains(t, err, "create spreadsheet API returned status 500")
}

func TestSheetsRepository_CreateSpreadsheet_WhenInvalidJSON_ShouldReturnError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer server.Close()

	repo := &sheetsRepository{baseURL: server.URL}
	id, err := repo.CreateSpreadsheet("test-token", "Mi Spreadsheet")

	assert.Empty(t, id)
	assert.ErrorContains(t, err, "parsing create spreadsheet response")
}

func TestSheetsRepository_CreateSpreadsheet_WhenServerDown_ShouldReturnError(t *testing.T) {
	repo := &sheetsRepository{baseURL: "http://localhost:1"}
	id, err := repo.CreateSpreadsheet("test-token", "Mi Spreadsheet")

	assert.Empty(t, id)
	assert.ErrorContains(t, err, "calling create spreadsheet API")
}
