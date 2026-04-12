package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/repository"
)

func NewConfigScreen(window fyne.Window, configRepo repository.ConfigRepository, onSave func()) fyne.CanvasObject {
	clientIDEntry := widget.NewEntry()
	clientIDEntry.SetPlaceHolder("Google Client ID")

	clientSecretEntry := widget.NewPasswordEntry()
	clientSecretEntry.SetPlaceHolder("Google Client Secret")

	portEntry := widget.NewEntry()
	portEntry.SetPlaceHolder("8081")

	spreadsheetEntry := widget.NewEntry()
	spreadsheetEntry.SetPlaceHolder("Spreadsheet ID")

	// Pre-llenar si ya existe config
	if cfg, err := configRepo.GetConfig(); err == nil {
		clientIDEntry.SetText(cfg.GoogleClientID)
		clientSecretEntry.SetText(cfg.GoogleClientSecret)
		portEntry.SetText(cfg.CallbackPort)
		spreadsheetEntry.SetText(cfg.SpreadsheetID)
	}

	saveBtn := widget.NewButton("Guardar", func() {
		port := portEntry.Text
		if port == "" {
			port = "8881"
		}

		err := configRepo.SetConfig(domain.Config{
			GoogleClientID:     clientIDEntry.Text,
			GoogleClientSecret: clientSecretEntry.Text,
			CallbackPort:       port,
			SpreadsheetID:      spreadsheetEntry.Text,
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
		widget.NewLabel("Spreadsheet ID"),
		spreadsheetEntry,
		widget.NewSeparator(),
		saveBtn,
	)

	return container.NewCenter(form)
}
