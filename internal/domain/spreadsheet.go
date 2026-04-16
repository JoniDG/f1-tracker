package domain

type SpreadsheetData struct {
	SpreadsheetId string `json:"spreadsheetId"`
	Properties    struct {
		Title         string `json:"title"`
		Locale        string `json:"locale"`
		AutoRecalc    string `json:"autoRecalc"`
		TimeZone      string `json:"timeZone"`
		DefaultFormat struct {
			BackgroundColor struct {
				Red   int `json:"red"`
				Green int `json:"green"`
				Blue  int `json:"blue"`
			} `json:"backgroundColor"`
			Padding struct {
				Top    int `json:"top"`
				Right  int `json:"right"`
				Bottom int `json:"bottom"`
				Left   int `json:"left"`
			} `json:"padding"`
			VerticalAlignment string `json:"verticalAlignment"`
			WrapStrategy      string `json:"wrapStrategy"`
			TextFormat        struct {
				ForegroundColor struct {
				} `json:"foregroundColor"`
				FontFamily           string `json:"fontFamily"`
				FontSize             int    `json:"fontSize"`
				Bold                 bool   `json:"bold"`
				Italic               bool   `json:"italic"`
				Strikethrough        bool   `json:"strikethrough"`
				Underline            bool   `json:"underline"`
				ForegroundColorStyle struct {
					RgbColor struct {
					} `json:"rgbColor"`
				} `json:"foregroundColorStyle"`
			} `json:"textFormat"`
			BackgroundColorStyle struct {
				RgbColor struct {
					Red   int `json:"red"`
					Green int `json:"green"`
					Blue  int `json:"blue"`
				} `json:"rgbColor"`
			} `json:"backgroundColorStyle"`
		} `json:"defaultFormat"`
		SpreadsheetTheme struct {
			PrimaryFontFamily string `json:"primaryFontFamily"`
			ThemeColors       []struct {
				ColorType string `json:"colorType"`
				Color     struct {
					RgbColor struct {
						Red   float64 `json:"red,omitempty"`
						Green float64 `json:"green,omitempty"`
						Blue  float64 `json:"blue,omitempty"`
					} `json:"rgbColor"`
				} `json:"color"`
			} `json:"themeColors"`
		} `json:"spreadsheetTheme"`
	} `json:"properties"`
	Sheets []SheetData `json:"sheets"`
}

type SheetData struct {
	Properties SheetDataProperties `json:"properties"`
}

type SheetDataProperties struct {
	SheetId        int    `json:"sheetId"`
	Title          string `json:"title"`
	Index          int    `json:"index"`
	SheetType      string `json:"sheetType"`
	GridProperties struct {
		RowCount    int `json:"rowCount"`
		ColumnCount int `json:"columnCount"`
	} `json:"gridProperties"`
}

type BatchUpdateRequest struct {
	Requests []BatchRequest `json:"requests"`
}

type BatchRequest struct {
	AddSheet *AddSheetRequest `json:"addSheet,omitempty"`
}

type AddSheetRequest struct {
	Properties AddSheetRequestProperties `json:"properties"`
}

type AddSheetRequestProperties struct {
	Title          string `json:"title"`
	GridProperties struct {
		RowCount    int `json:"rowCount"`
		ColumnCount int `json:"columnCount"`
	} `json:"gridProperties,omitempty"`
	TabColor struct {
		Red   float64 `json:"red"`
		Green float64 `json:"green"`
		Blue  float64 `json:"blue"`
	} `json:"tabColor,omitempty"`
}

type BatchResponse struct {
	SpreadsheetId string         `json:"spreadsheetId"`
	Replies       []BatchReplies `json:"replies"`
}

type BatchReplies struct {
	AddSheet *AddSheetResponse `json:"addSheet"`
}

type AddSheetResponse struct {
	Properties AddSheetResponseProperties `json:"properties"`
}
type AddSheetResponseProperties struct {
	SheetId        int    `json:"sheetId"`
	Title          string `json:"title"`
	Index          int    `json:"index"`
	SheetType      string `json:"sheetType"`
	GridProperties struct {
		RowCount    int `json:"rowCount"`
		ColumnCount int `json:"columnCount"`
	} `json:"gridProperties"`
	TabColorStyle struct {
		RgbColor struct {
		} `json:"rgbColor"`
	} `json:"tabColorStyle"`
}

type UpdateSheetValuesRequest struct {
	Range          string     `json:"range"`
	MajorDimension string     `json:"majorDimension"`
	Values         [][]string `json:"values"`
}

type UpdateSheetValuesResponse struct {
	SpreadsheetId  string `json:"spreadsheetId"`
	UpdatedRange   string `json:"updatedRange"`
	UpdatedRows    int    `json:"updatedRows"`
	UpdatedColumns int    `json:"updatedColumns"`
	UpdatedCells   int    `json:"updatedCells"`
}

type CreateSpreadsheetRequest struct {
	Properties CreateSpreadsheetRequestProperties `json:"properties"`
}

type CreateSpreadsheetRequestProperties struct {
	Title string `json:"title"`
}

type CreateSpreadsheetResponse struct {
	SpreadsheetId string `json:"spreadsheetId"`
}

type GetSheetValuesResponse struct {
	Range          string     `json:"range"`
	MajorDimension string     `json:"majorDimension"`
	Values         [][]string `json:"values"`
}
