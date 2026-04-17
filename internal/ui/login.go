package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

func NewLoginScreen(window fyne.Window, authSvc service.AuthService, onLogin func(*domain.User)) fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	loginBtn := widget.NewButton("Login con Google", nil)
	loginBtn.OnTapped = func() {
		loginBtn.Disable()
		statusLabel.SetText("Esperando autorizacion en el navegador...")
		progressBar.Show()

		go func() {
			user, err := authSvc.Login()
			if err != nil {
				fyne.Do(func() {
					progressBar.Hide()
					loginBtn.Enable()
					statusLabel.SetText("Error: " + err.Error())
				})
				return
			}
			fyne.Do(func() {
				progressBar.Hide()
				onLogin(user)
			})
		}()
	}

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabel("F1 Tracker"),
			widget.NewSeparator(),
			loginBtn,
			progressBar,
			statusLabel,
		),
	)
}
