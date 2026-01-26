package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
)

type StatusCard struct {
	widget.BaseWidget
	app          *models.App
	statusLabel  *widget.Label
	remoteLabel  *widget.Label
	localLabel   *widget.Label
	actionButton *widget.Button
}

func NewStatusCard(app *models.App) *StatusCard {
	card := &StatusCard{
		app:          app,
		statusLabel:  widget.NewLabel("Status: Loading..."),
		remoteLabel:  widget.NewLabel("Remote: ..."),
		localLabel:   widget.NewLabel("Local: ..."),
		actionButton: widget.NewButton("Deploy", func() {}),
	}
	card.ExtendBaseWidget(card)
	card.Refresh()
	return card
}

func (sc *StatusCard) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 240, A: 255})

	content := container.NewVBox(
		sc.statusLabel,
		sc.remoteLabel,
		sc.localLabel,
	)

	mainContent := container.NewBorder(
		nil, nil, nil, sc.actionButton,
		content,
	)

	padded := container.New(layout.NewPaddedLayout(), mainContent)

	return &statusCardRenderer{
		background: background,
		content:    padded,
		objects:    []fyne.CanvasObject{background, padded},
	}
}

type statusCardRenderer struct {
	background *canvas.Rectangle
	content    *fyne.Container
	objects    []fyne.CanvasObject
}

func (r *statusCardRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *statusCardRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

func (r *statusCardRenderer) Refresh() {
	r.background.Refresh()
	r.content.Refresh()
}

func (r *statusCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *statusCardRenderer) Destroy() {}
