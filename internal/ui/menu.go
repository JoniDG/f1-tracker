package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
)

func NewMenuScreen(window fyne.Window, user *domain.User, onMyTracks, onAddTime, onFriends func()) fyne.CanvasObject {
	userLabel := "F1 Tracker"
	if user != nil {
		userLabel = "Logueado como " + user.DisplayName
	}

	return container.NewCenter(
		container.NewVBox(
			widget.NewLabel("F1 Tracker"),
			widget.NewSeparator(),
			widget.NewLabel(userLabel),
			widget.NewSeparator(),
			widget.NewButton("Mis Tiempos", onMyTracks),
			widget.NewButton("Agregar / Actualizar Tiempo", onAddTime),
			widget.NewButton("Tiempos de Amigos", onFriends),
		),
	)
}
