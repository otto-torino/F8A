package utils

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var Scroll *container.Scroll

func MakeButton(text string, cb func()) *widget.Button {
	return widget.NewButton(text, cb)
}
