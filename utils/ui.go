package utils

import "fyne.io/fyne/v2/widget"

func MakeButton(text string, cb func()) *widget.Button {
	return widget.NewButton(text, cb)
}
