package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

// NewLoginScreen crea la pantalla de login con un boton "Login con Google".
// Recibe el window (para context), el authService (para ejecutar el login),
// y un callback onLogin que se llama cuando el login es exitoso con los datos del usuario.
func NewLoginScreen(window fyne.Window, authSvc service.AuthService, onLogin func(*domain.User)) fyne.CanvasObject {
	statusLabel := widget.NewLabel("")
	// ProgressBar infinita: se muestra mientras el usuario esta autorizando en el browser
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	loginBtn := widget.NewButton("Login con Google", nil)
	// OnTapped se ejecuta cuando el usuario hace click en el boton
	loginBtn.OnTapped = func() {
		// Deshabilita el boton para evitar clicks multiples mientras se autentica
		loginBtn.Disable()
		statusLabel.SetText("Esperando autorizacion en el navegador...")
		progressBar.Show()

		// Lanza el flujo de login en una goroutine para no bloquear la UI de Fyne.
		// authSvc.Login() es bloqueante: espera hasta que el usuario autorice en el browser.
		go func() {
			user, err := authSvc.Login()
			if err != nil {
				// fyne.Do ejecuta la funcion en el hilo principal de Fyne.
				// Es obligatorio para modificar widgets desde una goroutine (Fyne v2.7+).
				fyne.Do(func() {
					progressBar.Hide()
					loginBtn.Enable()
					statusLabel.SetText("Error: " + err.Error())
				})
				return
			}
			// Login exitoso: ocultamos el progress y navegamos a la siguiente pantalla
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
