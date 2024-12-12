package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/otto-torino/f8a/utils"
)

var mainContent *fyne.Container

func MakeMainContent() *fyne.Container {

	mainContent = container.New(layout.NewPaddedLayout())

	// default content
	HandleHomeSection()

	// react to events
	utils.Dispatcher.On(utils.AppSelect, func(args ...interface{}) {
		id := args[0].(int)
		HandleWebAppSection(id)
	})
	utils.Dispatcher.On(utils.AppDelete, func(args ...interface{}) {
		HandleHomeSection()
	})

	utils.Dispatcher.On(utils.AppUpdate, func(args ...interface{}) {
		HandleHomeSection()
	})

	return mainContent
}

func HandleHomeSection() {
	if mainContent != nil {
		mainContent.RemoveAll()
	}

	text1 := canvas.NewText("Otto Frontend Apps Manager", color.White)
	text1.TextSize = 18
	text2 := canvas.NewText("Developed by Otto srl", color.White)

	textContainer := container.NewVBox(text1, text2)
	mainContent.Add(textContainer)
}
