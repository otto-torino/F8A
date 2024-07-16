package utils

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

func MakeTitle(text string) *canvas.Text {
	title := canvas.NewText(text, theme.ForegroundColor())
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	return title
}
