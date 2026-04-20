package ui

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

var spreadsheetIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{25,}$`)

func NewSheetSetupScreen(window fyne.Window, authSvc service.AuthService, trackerSvc service.TrackerService, user *domain.User, onComplete func()) fyne.CanvasObject {
	contentBox := container.NewVBox()

	var render func()
	var onBack func()

	render = func() {
		contentBox.RemoveAll()
		cfg, err := authSvc.GetConfig()
		if err != nil {
			dialog.ShowError(fmt.Errorf("no se pudo leer la configuracion: %w", err), window)
			contentBox.Add(widget.NewLabel("Error al leer la configuracion."))
			contentBox.Add(widget.NewButton("Reintentar", render))
			contentBox.Refresh()
			return
		}
		if cfg == nil || cfg.SpreadsheetID == "" {
			showSpreadsheetConnectionUI(window, authSvc, trackerSvc, user, contentBox, onBack, onComplete)
		} else {
			showUsernameUI(window, trackerSvc, user, contentBox, false, cfg.Username, onBack, onComplete)
		}
		contentBox.Refresh()
	}

	onBack = func() {
		cfg, err := authSvc.GetConfig()
		if err == nil && cfg != nil {
			cfg.SpreadsheetID = ""
			_ = authSvc.SetConfig(*cfg)
		}
		render()
	}

	render()

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabel("Configuracion de Spreadsheet"),
			widget.NewSeparator(),
			contentBox,
		),
	)
}

func showSpreadsheetConnectionUI(window fyne.Window, authSvc service.AuthService, trackerSvc service.TrackerService, user *domain.User, contentBox *fyne.Container, onBack, onComplete func()) {
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
				showUsernameUI(window, trackerSvc, user, contentBox, true, "", onBack, onComplete)
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
			existingUsername := ""
			if cfg, cfgErr := authSvc.GetConfig(); cfgErr == nil && cfg != nil {
				existingUsername = cfg.Username
			}
			fyne.Do(func() {
				progressBar.Hide()
				contentBox.RemoveAll()
				showUsernameUI(window, trackerSvc, user, contentBox, false, existingUsername, onBack, onComplete)
			})
		}()
	}

	contentBox.RemoveAll()
	contentBox.Add(radioGroup)
	contentBox.Add(createSection)
	contentBox.Add(connectSection)
	contentBox.Add(progressBar)
}

func showUsernameUI(window fyne.Window, trackerSvc service.TrackerService, user *domain.User, contentBox *fyne.Container, isNewSpreadsheet bool, existingUsername string, onBack, onComplete func()) {
	reuseMode := existingUsername != "" && !isNewSpreadsheet

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Tu nombre en el spreadsheet")
	if reuseMode {
		usernameEntry.SetText(existingUsername)
		usernameEntry.Disable()
	} else if user != nil && user.DisplayName != "" {
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

		email := ""
		if user != nil {
			email = user.Email
		}

		go func() {
			if reuseMode {
				status, err := trackerSvc.VerifySheetOwnership(name, email)
				if err != nil {
					fyne.Do(func() {
						progressBar.Hide()
						confirmBtn.Enable()
						dialog.ShowError(err, window)
					})
					return
				}
				switch status {
				case service.SheetOwnershipOwned:
					fyne.Do(func() {
						progressBar.Hide()
						onComplete()
					})
				case service.SheetOwnershipForeign:
					fyne.Do(func() {
						progressBar.Hide()
						confirmBtn.Enable()
						dialog.ShowError(fmt.Errorf("la hoja '%s' pertenece a otro usuario (su email no coincide con el tuyo)", name), window)
					})
				case service.SheetOwnershipUnprotected:
					fyne.Do(func() {
						progressBar.Hide()
						confirmBtn.Enable()
						dialog.ShowError(fmt.Errorf("la hoja '%s' es una hoja legacy sin proteccion. Borrala manualmente en Google Sheets y recreala desde la app", name), window)
					})
				case service.SheetOwnershipMissing:
					fyne.Do(func() {
						progressBar.Hide()
						confirmBtn.Enable()
						dialog.ShowError(fmt.Errorf("la hoja '%s' no existe en esta spreadsheet. Volve a elegir otra o borra el Username del config para crear una nueva", name), window)
					})
				}
				return
			}

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
			err = trackerSvc.SetupUser(name, isNewSpreadsheet, email)
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
	if onBack != nil {
		contentBox.Add(widget.NewButton("Volver", onBack))
	}
	contentBox.Add(progressBar)
}

func parseSpreadsheetID(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err != nil || u.Host != "docs.google.com" {
			return ""
		}
		parts := strings.Split(u.Path, "/d/")
		if len(parts) < 2 {
			return ""
		}
		id := parts[1]
		if idx := strings.Index(id, "/"); idx != -1 {
			id = id[:idx]
		}
		if spreadsheetIDRegex.MatchString(id) {
			return id
		}
		return ""
	}
	if spreadsheetIDRegex.MatchString(input) {
		return input
	}
	return ""
}
