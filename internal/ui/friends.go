package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/service"
)

func NewFriendsScreen(window fyne.Window, trackerSvc service.TrackerService, onBack func()) fyne.CanvasObject {
	backBtn := widget.NewButton("Volver", onBack)
	tableContainer := container.NewStack()
	statusLabel := widget.NewLabel("Cargando lista de amigos...")

	friendSelect := widget.NewSelect([]string{}, func(selected string) {
		statusLabel.SetText("Cargando tiempos de " + selected + "...")
		tableContainer.Objects = []fyne.CanvasObject{statusLabel}
		tableContainer.Refresh()

		go func() {
			tracks, err := trackerSvc.GetFriendTracks(selected)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
				table := buildTrackTable(tracks)
				tableContainer.Objects = []fyne.CanvasObject{table}
				tableContainer.Refresh()
			})
		}()
	})
	friendSelect.PlaceHolder = "Selecciona un amigo"
	friendSelect.Disable()

	tableContainer.Add(statusLabel)

	go func() {
		friends, err := trackerSvc.GetFriendsList()
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, window)
				statusLabel.SetText("Error al cargar amigos")
				return
			}
			if len(friends) == 0 {
				statusLabel.SetText("No hay amigos en esta spreadsheet")
				return
			}
			friendSelect.Options = friends
			friendSelect.Enable()
			statusLabel.SetText("Selecciona un amigo para ver sus tiempos")
		})
	}()

	top := container.NewVBox(
		widget.NewLabel("Tiempos de Amigos"),
		widget.NewSeparator(),
		friendSelect,
	)

	return container.NewBorder(top, backBtn, nil, nil, tableContainer)
}
