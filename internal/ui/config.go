package ui

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

const googleClientIDSuffix = ".apps.googleusercontent.com"

func NewConfigScreen(window fyne.Window, authSvc service.AuthService, onSave func()) fyne.CanvasObject {
	clientIDEntry := widget.NewEntry()
	clientIDEntry.SetPlaceHolder("Google Client ID")

	clientSecretEntry := widget.NewPasswordEntry()
	clientSecretEntry.SetPlaceHolder("Google Client Secret")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("8081")

	spreadsheetEntry := widget.NewEntry()
	spreadsheetEntry.SetPlaceHolder("Spreadsheet ID")

	if cfg, err := authSvc.GetConfig(); err == nil {
		clientIDEntry.SetText(cfg.GoogleClientID)
		clientSecretEntry.SetText(cfg.GoogleClientSecret)
		portEntry.SetText(cfg.CallbackPort)
		spreadsheetEntry.SetText(cfg.SpreadsheetID)
	}

	saveBtn := widget.NewButton("Guardar", func() {
		clientID := strings.TrimSpace(clientIDEntry.Text)
		if clientID == "" {
			dialog.ShowError(fmt.Errorf("el Google Client ID es obligatorio"), window)
			return
		}
		if !strings.HasSuffix(clientID, googleClientIDSuffix) {
			dialog.ShowError(fmt.Errorf("el Google Client ID debe terminar en %s", googleClientIDSuffix), window)
			return
		}

		clientSecret := strings.TrimSpace(clientSecretEntry.Text)
		if clientSecret == "" {
			dialog.ShowError(fmt.Errorf("el Google Client Secret es obligatorio"), window)
			return
		}

		port := strings.TrimSpace(portEntry.Text)
		if port == "" {
			port = "8081"
		} else {
			portNum, err := strconv.Atoi(port)
			if err != nil || portNum < 1 || portNum > 65535 {
				dialog.ShowError(fmt.Errorf("el puerto debe ser un numero entre 1 y 65535"), window)
				return
			}
		}

		spreadsheetID := strings.TrimSpace(spreadsheetEntry.Text)
		if spreadsheetID != "" {
			parsed := parseSpreadsheetID(spreadsheetID)
			if parsed == "" {
				dialog.ShowError(fmt.Errorf("el Spreadsheet ID no es valido"), window)
				return
			}
			spreadsheetID = parsed
		}

		err := authSvc.SetConfig(domain.Config{
			GoogleClientID:     clientID,
			GoogleClientSecret: clientSecret,
			CallbackPort:       port,
			SpreadsheetID:      spreadsheetID,
		})
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		onSave()
	})

	form := container.NewVBox(
		widget.NewLabel("Configuracion"),
		widget.NewSeparator(),
		widget.NewLabel("Google Client ID"),
		clientIDEntry,
		widget.NewLabel("Google Client Secret"),
		clientSecretEntry,
		widget.NewLabel("Puerto callback"),
		portEntry,
		widget.NewLabel("Spreadsheet ID (opcional)"),
		spreadsheetEntry,
		widget.NewSeparator(),
		saveBtn,
	)

	return container.NewCenter(form)
}
