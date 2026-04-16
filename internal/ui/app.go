package ui

import (
	"log"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

// FyneApp es el orquestador principal de la aplicacion.
// Maneja la ventana de Fyne y la navegacion entre pantallas (config -> login -> success).
type FyneApp struct {
	app        fyne.App
	window     fyne.Window
	authSvc    service.AuthService
	trackerSvc service.TrackerService
}

// NewFyneApp crea la aplicacion Fyne con una ventana de 500x400.
func NewFyneApp(authSvc service.AuthService, trackerSvc service.TrackerService) *FyneApp {
	// app.New() inicializa el framework Fyne (solo se llama una vez en toda la app)
	a := fyneApp.New()
	// NewWindow crea la ventana principal con el titulo dado
	w := a.NewWindow("F1 Tracker")
	w.Resize(fyne.NewSize(500, 400))

	return &FyneApp{
		app:        a,
		window:     w,
		authSvc:    authSvc,
		trackerSvc: trackerSvc,
	}
}

// Run decide que pantalla mostrar al inicio segun el estado de la config y el token,
// y arranca el loop principal de Fyne (ShowAndRun es bloqueante, no retorna hasta cerrar la ventana).
func (fa *FyneApp) Run() {
	// Si no hay config o las credenciales son invalidas -> pantalla de configuracion
	if !fa.authSvc.HasValidConfig() {
		fa.showConfigScreen()
	} else if !fa.authSvc.HasStoredToken() {
		// Si hay config pero no hay token guardado -> pantalla de login
		fa.showLoginScreen()
	} else {
		// Si hay config y token -> obtener info del usuario.
		// TrackerService se encarga de validar/refrescar el token internamente.
		user, err := fa.trackerSvc.GetCurrentUser()
		if err != nil {
			// Si falla (refresh token invalido, etc.) -> mandar al login
			log.Printf("Error fetching user info: %v", err)
			fa.showLoginScreen()
		} else {
			fa.showSuccessScreen(user)
		}
	}

	// ShowAndRun muestra la ventana y arranca el event loop de Fyne.
	// Este metodo bloquea hasta que el usuario cierre la ventana.
	fa.window.ShowAndRun()
}

// showConfigScreen muestra la pantalla de configuracion.
// Al guardar la config exitosamente, navega automaticamente a la pantalla de login.
func (fa *FyneApp) showConfigScreen() {
	// window.SetContent reemplaza todo el contenido de la ventana con el nuevo widget.
	// Asi se implementa la "navegacion" entre pantallas en Fyne.
	fa.window.SetContent(NewConfigScreen(fa.window, fa.authSvc, fa.showLoginScreen))
}

// showLoginScreen muestra la pantalla de login con el boton "Login con Google".
// Al loguearse exitosamente, navega a la pantalla de exito.
func (fa *FyneApp) showLoginScreen() {
	fa.window.SetContent(NewLoginScreen(fa.window, fa.authSvc, fa.showSuccessScreen))
}

// showSuccessScreen muestra un mensaje de exito con el nombre del usuario logueado.
// Si user es nil (caso: ya estaba logueado al iniciar), muestra un mensaje generico.
func (fa *FyneApp) showSuccessScreen(user *domain.User) {
	msg := "Ya estas logueado"
	if user != nil {
		msg = "Logueado como " + user.DisplayName + " (" + user.Email + ")"
	}

	fa.window.SetContent(
		container.NewCenter(
			container.NewVBox(
				widget.NewLabel(msg),
			),
		),
	)

	err := fa.trackerSvc.EnsureSheetExists("JoniDG")
	if err != nil {
		log.Printf("Error: %v", err)
	}

	/*err := fa.trackerSvc.GetSheet()
	 */
}
