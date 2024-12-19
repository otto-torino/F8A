package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/theme"
	"github.com/otto-torino/f8a/utils"
)

var navContent *fyne.Container

func MakeSidebar(addCb func()) *fyne.Container {
	// get theme variant
	registry := utils.Registry()
	themeVariant := (*registry.Application).Settings().ThemeVariant()
	t := theme.F8aTheme{}

	title := canvas.NewText("Apps", color.RGBA{R: 255, G: 153, B: 0, A: 255})
	title.TextSize = 18
	titleContainer := container.New(layout.NewVBoxLayout(), title)

	navContent = container.New(layout.NewStackLayout())
	UpdateNavContent()

	background := canvas.NewRectangle(t.SidebarBg(themeVariant))
	sidebar := container.New(layout.NewStackLayout(), background, container.NewPadded(container.NewBorder(titleContainer, nil, nil, nil, navContent)))

	utils.Dispatcher.On(utils.AppChange, func(args ...interface{}) {
		UpdateNavContent()
	})

	utils.Dispatcher.On(utils.AppDelete, func(args ...interface{}) {
		UpdateNavContent()
	})

	return sidebar
}

func UpdateNavContent() {
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

	navContent.RemoveAll()
	navContent.Add(list)
	navContent.Refresh()
}
