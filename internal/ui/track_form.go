package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/defines"
	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

func NewTrackFormScreen(window fyne.Window, trackerSvc service.TrackerService, onBack func()) fyne.CanvasObject {
	bestLapEntry := widget.NewEntry()
	bestLapEntry.SetPlaceHolder("1:23.456")

	bestS1Entry := widget.NewEntry()
	bestS1Entry.SetPlaceHolder("S1")

	bestS2Entry := widget.NewEntry()
	bestS2Entry.SetPlaceHolder("S2")

	bestS3Entry := widget.NewEntry()
	bestS3Entry.SetPlaceHolder("S3")

	lapS1Entry := widget.NewEntry()
	lapS1Entry.SetPlaceHolder("S1 Vuelta")

	lapS2Entry := widget.NewEntry()
	lapS2Entry.SetPlaceHolder("S2 Vuelta")

	lapS3Entry := widget.NewEntry()
	lapS3Entry.SetPlaceHolder("S3 Vuelta")

	carEntry := widget.NewEntry()
	carEntry.SetPlaceHolder("Auto")

	dateEntry := widget.NewEntry()
	dateEntry.SetText(time.Now().Format("2006-01-02"))

	entries := []*widget.Entry{bestLapEntry, bestS1Entry, bestS2Entry, bestS3Entry, lapS1Entry, lapS2Entry, lapS3Entry, carEntry, dateEntry}

	var selectedTrack string
	trackSelect := widget.NewSelect(defines.Tracks, func(selected string) {
		selectedTrack = selected
		go func() {
			tracks, err := trackerSvc.GetMyTracks()
			fyne.Do(func() {
				if err != nil {
					for _, e := range entries[:8] {
						e.SetText("")
					}
					dialog.ShowError(fmt.Errorf("no se pudieron cargar los tiempos existentes: %w", err), window)
					return
				}
				for _, t := range tracks {
					if t.TrackName == selected {
						bestLapEntry.SetText(t.BestLapTime)
						bestS1Entry.SetText(t.BestS1)
						bestS2Entry.SetText(t.BestS2)
						bestS3Entry.SetText(t.BestS3)
						lapS1Entry.SetText(t.LapS1)
						lapS2Entry.SetText(t.LapS2)
						lapS3Entry.SetText(t.LapS3)
						carEntry.SetText(t.Car)
						if t.Date != "" {
							dateEntry.SetText(t.Date)
						}
						return
					}
				}
				for _, e := range entries[:8] {
					e.SetText("")
				}
			})
		}()
	})
	trackSelect.PlaceHolder = "Selecciona un circuito"

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Hide()

	saveBtn := widget.NewButton("Guardar", nil)
	saveBtn.OnTapped = func() {
		if selectedTrack == "" {
			dialog.ShowError(fmt.Errorf("selecciona un circuito"), window)
			return
		}
		if bestLapEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("ingresa el mejor tiempo de vuelta"), window)
			return
		}

		track := domain.TrackTime{
			TrackName:   selectedTrack,
			BestLapTime: bestLapEntry.Text,
			BestS1:      bestS1Entry.Text,
			BestS2:      bestS2Entry.Text,
			BestS3:      bestS3Entry.Text,
			LapS1:       lapS1Entry.Text,
			LapS2:       lapS2Entry.Text,
			LapS3:       lapS3Entry.Text,
			Car:         carEntry.Text,
			Date:        dateEntry.Text,
		}

		saveBtn.Disable()
		progressBar.Show()

		go func() {
			err := trackerSvc.SaveTrackTime(track)
			fyne.Do(func() {
				progressBar.Hide()
				if err != nil {
					saveBtn.Enable()
					dialog.ShowError(err, window)
					return
				}
				dialog.ShowInformation("Guardado", "Tiempo guardado correctamente", window)
				onBack()
			})
		}()
	}

	cancelBtn := widget.NewButton("Cancelar", onBack)

	form := container.NewVBox(
		widget.NewLabel("Agregar / Actualizar Tiempo"),
		widget.NewSeparator(),
		widget.NewLabel("Circuito"),
		trackSelect,
		widget.NewLabel("Mejor Vuelta"),
		bestLapEntry,
		widget.NewLabel("Mejor S1"),
		bestS1Entry,
		widget.NewLabel("Mejor S2"),
		bestS2Entry,
		widget.NewLabel("Mejor S3"),
		bestS3Entry,
		widget.NewLabel("S1 Vuelta"),
		lapS1Entry,
		widget.NewLabel("S2 Vuelta"),
		lapS2Entry,
		widget.NewLabel("S3 Vuelta"),
		lapS3Entry,
		widget.NewLabel("Auto"),
		carEntry,
		widget.NewLabel("Fecha"),
		dateEntry,
		widget.NewSeparator(),
		container.NewHBox(saveBtn, cancelBtn),
		progressBar,
	)

	return container.NewVScroll(form)
}
