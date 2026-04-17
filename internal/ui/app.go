package ui

import (
	"log"

	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

// FyneApp es el orquestador principal de la aplicacion.
// Maneja la ventana de Fyne y la navegacion entre pantallas.
type FyneApp struct {
	app        fyne.App
	window     fyne.Window
	authSvc    service.AuthService
	trackerSvc service.TrackerService
}

// NewFyneApp crea la aplicacion Fyne con una ventana de 500x400.
func NewFyneApp(authSvc service.AuthService, trackerSvc service.TrackerService) *FyneApp {
	a := fyneApp.New()
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
	if !fa.authSvc.HasValidConfig() {
		fa.showConfigScreen()
	} else if !fa.authSvc.HasStoredToken() {
		fa.showLoginScreen()
	} else {
		user, err := fa.trackerSvc.GetCurrentUser()
		if err != nil {
			log.Printf("Error fetching user info: %v", err)
			fa.showLoginScreen()
		} else {
			fa.showPostLoginScreen(user)
		}
	}

	fa.window.ShowAndRun()
}

// showConfigScreen muestra la pantalla de configuracion.
func (fa *FyneApp) showConfigScreen() {
	fa.window.SetContent(NewConfigScreen(fa.window, fa.authSvc, fa.showLoginScreen))
}

// showLoginScreen muestra la pantalla de login con el boton "Login con Google".
func (fa *FyneApp) showLoginScreen() {
	fa.window.SetContent(NewLoginScreen(fa.window, fa.authSvc, fa.showPostLoginScreen))
}

// showPostLoginScreen decide si mostrar sheet setup o el menu principal.
func (fa *FyneApp) showPostLoginScreen(user *domain.User) {
	if fa.trackerSvc.NeedsSheetSetup() {
		fa.showSheetSetupScreen(user)
	} else {
		fa.showMenuScreen(user)
	}
}

// showSheetSetupScreen muestra la pantalla de configuracion de spreadsheet y username.
func (fa *FyneApp) showSheetSetupScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 500))
	fa.window.SetContent(NewSheetSetupScreen(fa.window, fa.authSvc, fa.trackerSvc, user, func() {
		fa.showMenuScreen(user)
	}))
}

// showMenuScreen muestra el menu principal con las opciones de navegacion.
func (fa *FyneApp) showMenuScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 400))
	fa.window.SetContent(NewMenuScreen(fa.window, user,
		func() { fa.showTracksScreen(user) },
		func() { fa.showTrackFormScreen(user) },
		func() { fa.showFriendsScreen(user) },
	))
}

// showTracksScreen muestra la tabla de tiempos del usuario.
func (fa *FyneApp) showTracksScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(1100, 600))
	fa.window.SetContent(NewTracksScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}

// showTrackFormScreen muestra el formulario para agregar/actualizar tiempos.
func (fa *FyneApp) showTrackFormScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 600))
	fa.window.SetContent(NewTrackFormScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}

// showFriendsScreen muestra la pantalla de tiempos de amigos.
func (fa *FyneApp) showFriendsScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(1100, 600))
	fa.window.SetContent(NewFriendsScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}
