package ui

import (
	"fyne.io/fyne/v2"
	fyneApp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

type FyneApp struct {
	app        fyne.App
	window     fyne.Window
	authSvc    service.AuthService
	trackerSvc service.TrackerService
}

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

func (fa *FyneApp) Run() {
	if !fa.authSvc.HasValidConfig() {
		fa.showConfigScreen()
	} else if !fa.authSvc.HasStoredToken() {
		fa.showLoginScreen()
	} else {
		user, err := fa.trackerSvc.GetCurrentUser()
		if err != nil {
			fa.showLoginScreen()
			dialog.ShowError(err, fa.window)
		} else {
			fa.showPostLoginScreen(user)
		}
	}

	fa.window.ShowAndRun()
}

func (fa *FyneApp) showConfigScreen() {
	fa.window.SetContent(NewConfigScreen(fa.window, fa.authSvc, fa.showLoginScreen))
}

func (fa *FyneApp) showLoginScreen() {
	fa.window.SetContent(NewLoginScreen(fa.window, fa.authSvc, fa.showPostLoginScreen))
}

func (fa *FyneApp) showPostLoginScreen(user *domain.User) {
	if fa.trackerSvc.NeedsSheetSetup() {
		fa.showSheetSetupScreen(user)
	} else {
		fa.showMenuScreen(user)
	}
}

func (fa *FyneApp) showSheetSetupScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 500))
	fa.window.SetContent(NewSheetSetupScreen(fa.window, fa.authSvc, fa.trackerSvc, user, func() {
		fa.showMenuScreen(user)
	}))
}

func (fa *FyneApp) showMenuScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 400))
	fa.window.SetContent(NewMenuScreen(fa.window, user,
		func() { fa.showTracksScreen(user) },
		func() { fa.showTrackFormScreen(user) },
		func() { fa.showFriendsScreen(user) },
	))
}

func (fa *FyneApp) showTracksScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(1100, 600))
	fa.window.SetContent(NewTracksScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}

func (fa *FyneApp) showTrackFormScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(500, 600))
	fa.window.SetContent(NewTrackFormScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}

func (fa *FyneApp) showFriendsScreen(user *domain.User) {
	fa.window.Resize(fyne.NewSize(1100, 600))
	fa.window.SetContent(NewFriendsScreen(fa.window, fa.trackerSvc, func() {
		fa.showMenuScreen(user)
	}))
}
