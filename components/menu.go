package components

import (
	"fyne.io/fyne/v2"
	"github.com/otto-torino/f8a/utils"
)

func MakeMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	newItem := fyne.NewMenuItem("New", nil)
	appItem := fyne.NewMenuItem("App", func() {
		utils.Dispatcher.Emit(utils.AppSelect, 0)
		HandleAddWebApp()
	})
	newItem.ChildMenu = fyne.NewMenu("",
		appItem,
	)

	// a quit item will be appended to our first (File) menu
	file := fyne.NewMenu("File", newItem)
	main := fyne.NewMainMenu(
		file,
	)
	return main
}
