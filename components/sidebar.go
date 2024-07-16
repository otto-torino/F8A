package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)

var sidebarContent *fyne.Container

func MakeSidebar(addCb func()) *fyne.Container {
	title := canvas.NewText("Apps", color.RGBA{R: 255, G: 153, B: 0, A: 255})
	title.TextSize = 18
	titleContainer := container.New(layout.NewVBoxLayout(), title)

	// addButton := widget.NewButton("Add App", func() {
	// 	utils.Dispatcher.Emit(utils.AppSelect, 0)
	// 	addCb()
	// })

	sidebarContent = container.New(layout.NewStackLayout())
	UpdateSidebarContent()

	// background := canvas.NewRectangle(color.RGBA{R: 33, G: 33, B: 33, A: 255})
	// background := canvas.NewRectangle(color.RGBA{R: 33, G: 33, B: 33, A: 255})
	sidebar := container.NewPadded(container.NewBorder(titleContainer, nil, nil, nil, sidebarContent))

	utils.Dispatcher.On(utils.AppChange, func(args ...interface{}) {
		UpdateSidebarContent()
	})

	utils.Dispatcher.On(utils.AppDelete, func(args ...interface{}) {
		UpdateSidebarContent()
	})

	return sidebar
}

func UpdateSidebarContent() {
	webapps, err := models.GetApps()
	if err != nil {
		return
	}

	list := widget.NewList(
		func() int {
			return len(webapps)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(webapps[i].Name)
		})
	list.OnSelected = func(id widget.ListItemID) {
		utils.Dispatcher.Emit(utils.AppSelect, webapps[id].ID)
	}
	utils.Dispatcher.On(utils.AppAdd, func(args ...interface{}) {
		list.UnselectAll()
		id := args[0].(int)
		if id == 0 {
			return
		}
		index := 0
		for i := range webapps {
			if webapps[i].ID == id {
				index = i
				break
			}
		}
		list.Select(index)
	})

	sidebarContent.RemoveAll()
	sidebarContent.Add(list)
	sidebarContent.Refresh()
}
