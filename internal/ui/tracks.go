package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/JoniDG/f1-tracker/internal/domain"
	"github.com/JoniDG/f1-tracker/internal/service"
)

var trackTableHeaders = []string{
	"Circuito", "Mejor Vuelta", "Mejor S1", "Mejor S2", "Mejor S3",
	"S1 Vuelta", "S2 Vuelta", "S3 Vuelta", "Auto", "Fecha",
}

func NewTracksScreen(window fyne.Window, trackerSvc service.TrackerService, onBack func()) fyne.CanvasObject {
	progressBar := widget.NewProgressBarInfinite()
	backBtn := widget.NewButton("Volver", onBack)

	content := container.NewBorder(nil, backBtn, nil, nil, progressBar)

	go func() {
		tracks, err := trackerSvc.GetMyTracks()
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, window)
				content.Objects = []fyne.CanvasObject{backBtn}
				content.Refresh()
				return
			}
			table := buildTrackTable(tracks)
			content.Objects = []fyne.CanvasObject{table, backBtn}
			content.Refresh()
		})
	}()

	return content
}

func buildTrackTable(tracks []domain.TrackTime) *widget.Table {
	table := widget.NewTable(
		func() (int, int) {
			return len(tracks) + 1, len(trackTableHeaders)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder text here")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)
			if id.Row == 0 {
				label.SetText(trackTableHeaders[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				return
			}
			label.TextStyle = fyne.TextStyle{}
			track := tracks[id.Row-1]
			label.SetText(getTrackField(track, id.Col))
		},
	)

	colWidths := []float32{140, 100, 80, 80, 80, 80, 80, 80, 120, 100}
	for i, w := range colWidths {
		table.SetColumnWidth(i, w)
	}

	return table
}

func getTrackField(track domain.TrackTime, col int) string {
	switch col {
	case 0:
		return track.TrackName
	case 1:
		return track.BestLapTime
	case 2:
		return track.BestS1
	case 3:
		return track.BestS2
	case 4:
		return track.BestS3
	case 5:
		return track.LapS1
	case 6:
		return track.LapS2
	case 7:
		return track.LapS3
	case 8:
		return track.Car
	case 9:
		return track.Date
	default:
		return ""
	}
}
