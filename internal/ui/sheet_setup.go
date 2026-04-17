package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

func NewSheetSetupScreen(window fyne.Window, authSvc service.AuthService, trackerSvc service.TrackerService, user *domain.User, onComplete func()) fyne.CanvasObject {
	contentBox := container.NewVBox()

	cfg, _ := authSvc.GetConfig()
	if cfg == nil || cfg.SpreadsheetID == "" {
		showSpreadsheetConnectionUI(window, trackerSvc, user, contentBox, onComplete)
	} else {
		showUsernameUI(window, trackerSvc, user, contentBox, onComplete)
	}

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabel("Configuracion de Spreadsheet"),
			widget.NewSeparator(),
			contentBox,
		),
	)
}

func showSpreadsheetConnectionUI(window fyne.Window, trackerSvc service.TrackerService, user *domain.User, contentBox *fyne.Container, onComplete func()) {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Nombre de la spreadsheet")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("URL o ID de la spreadsheet")

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	createBtn := widget.NewButton("Crear", nil)
	connectBtn := widget.NewButton("Conectar", nil)

	createSection := container.NewVBox(
		widget.NewLabel("Nombre:"),
		titleEntry,
		createBtn,
	)

	connectSection := container.NewVBox(
		widget.NewLabel("URL o ID:"),
		urlEntry,
		connectBtn,
	)

	createSection.Hide()
	connectSection.Hide()

	radioGroup := widget.NewRadioGroup([]string{"Crear nueva spreadsheet", "Usar spreadsheet existente"}, func(selected string) {
		switch selected {
		case "Crear nueva spreadsheet":
			createSection.Show()
			connectSection.Hide()
		case "Usar spreadsheet existente":
			createSection.Hide()
			connectSection.Show()
		}
	})

	createBtn.OnTapped = func() {
		if titleEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("ingresa un nombre para la spreadsheet"), window)
			return
		}
		createBtn.Disable()
		progressBar.Show()

		go func() {
			id, err := trackerSvc.CreateSpreadsheet(titleEntry.Text)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					createBtn.Enable()
					dialog.ShowError(err, window)
				})
				return
			}
			err = trackerSvc.SaveSpreadsheetID(id)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					createBtn.Enable()
					dialog.ShowError(err, window)
				})
				return
			}
			fyne.Do(func() {
				progressBar.Hide()
				contentBox.RemoveAll()
				showUsernameUI(window, trackerSvc, user, contentBox, onComplete)
			})
		}()
	}

	connectBtn.OnTapped = func() {
		if urlEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("ingresa la URL o ID de la spreadsheet"), window)
			return
		}
		id := parseSpreadsheetID(urlEntry.Text)
		if id == "" {
			dialog.ShowError(fmt.Errorf("no se pudo extraer el ID de la spreadsheet"), window)
			return
		}
		connectBtn.Disable()
		progressBar.Show()

		go func() {
			err := trackerSvc.SaveSpreadsheetID(id)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					connectBtn.Enable()
					dialog.ShowError(err, window)
				})
				return
			}
			fyne.Do(func() {
				progressBar.Hide()
				contentBox.RemoveAll()
				showUsernameUI(window, trackerSvc, user, contentBox, onComplete)
			})
		}()
	}

	contentBox.RemoveAll()
	contentBox.Add(radioGroup)
	contentBox.Add(createSection)
	contentBox.Add(connectSection)
	contentBox.Add(progressBar)
}

func showUsernameUI(window fyne.Window, trackerSvc service.TrackerService, user *domain.User, contentBox *fyne.Container, onComplete func()) {
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Tu nombre en el spreadsheet")
	if user != nil && user.DisplayName != "" {
		usernameEntry.SetText(user.DisplayName)
	}

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	confirmBtn := widget.NewButton("Confirmar", nil)
	confirmBtn.OnTapped = func() {
		name := strings.TrimSpace(usernameEntry.Text)
		if name == "" {
			dialog.ShowError(fmt.Errorf("ingresa un nombre de usuario"), window)
			return
		}
		confirmBtn.Disable()
		progressBar.Show()

		go func() {
			available, err := trackerSvc.IsUsernameAvailable(name)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					confirmBtn.Enable()
					dialog.ShowError(err, window)
				})
				return
			}
			if !available {
				fyne.Do(func() {
					progressBar.Hide()
					confirmBtn.Enable()
					dialog.ShowError(fmt.Errorf("el nombre '%s' ya esta en uso", name), window)
				})
				return
			}
			err = trackerSvc.SetupUser(name)
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					confirmBtn.Enable()
					dialog.ShowError(err, window)
				})
				return
			}
			fyne.Do(func() {
				progressBar.Hide()
				onComplete()
			})
		}()
	}

	contentBox.Add(widget.NewLabel("Nombre de usuario"))
	contentBox.Add(usernameEntry)
	contentBox.Add(confirmBtn)
	contentBox.Add(progressBar)
}

func parseSpreadsheetID(input string) string {
	input = strings.TrimSpace(input)
	if strings.Contains(input, "docs.google.com/spreadsheets/d/") {
		parts := strings.Split(input, "/d/")
		if len(parts) < 2 {
			return ""
		}
		id := parts[1]
		if idx := strings.Index(id, "/"); idx != -1 {
			id = id[:idx]
		}
		if idx := strings.Index(id, "?"); idx != -1 {
			id = id[:idx]
		}
		return id
	}
	return input
}
