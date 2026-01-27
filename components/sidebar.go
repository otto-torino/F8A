// Package components
package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/theme"
	"github.com/otto-torino/f8a/utils"
)

var navContent *fyne.Container
var sidebarItems []*SidebarItem
var currentSelectedID int
var currentThemeVariant fyne.ThemeVariant

func MakeSidebar(addCb func()) *fyne.Container {
	registry := utils.Registry()
	currentThemeVariant = (*registry.Application).Settings().ThemeVariant()
	t := theme.F8aTheme{}

	// Title
	title := canvas.NewText("Apps", color.RGBA{R: 255, G: 153, B: 0, A: 255})
	title.TextSize = 18

	// Separator
	separator := canvas.NewRectangle(t.SidebarSeparator(currentThemeVariant))
	separator.SetMinSize(fyne.NewSize(0, 1))
	separatorContainer := container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(),
		separator,
		layout.NewSpacer(),
	)

	titleContainer := container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(),
		title,
		layout.NewSpacer(),
		separatorContainer,
	)

	// Nav content
	navContent = container.New(layout.NewVBoxLayout())
	UpdateNavContent()

	// Scrollable list
	scrollContent := container.NewVScroll(navContent)

	// Main content with padding
	mainContent := container.NewBorder(titleContainer, nil, nil, nil, scrollContent)
	paddedContent := container.NewPadded(mainContent)

	// Background and border
	background := canvas.NewRectangle(t.SidebarBg(currentThemeVariant))
	border := canvas.NewRectangle(t.SidebarBorder(currentThemeVariant))
	border.SetMinSize(fyne.NewSize(1, 0))

	borderContainer := container.New(layout.NewBorderLayout(nil, nil, nil, border), border)

	sidebar := container.New(layout.NewStackLayout(),
		background,
		borderContainer,
		paddedContent,
	)

	utils.Dispatcher.On(utils.AppChange, func(args ...any) {
		UpdateNavContent()
	})

	utils.Dispatcher.On(utils.AppDelete, func(args ...any) {
		UpdateNavContent()
	})

	utils.Dispatcher.On(utils.AppAdd, func(args ...any) {
		id := args[0].(int)

		// Re-fetch webapps to get the latest data
		webapps, err := models.GetApps()
		if err != nil {
			return
		}

		if id == 0 {
			// Deselect all
			for _, item := range sidebarItems {
				item.SetSelected(false)
			}
			currentSelectedID = 0
			return
		}

		// Find and select the new app
		for i, app := range webapps {
			if app.ID == id {
				updateSelection(i)
				break
			}
		}
	})

	return sidebar
}

func UpdateNavContent() {
	webapps, err := models.GetApps()
	if err != nil {
		return
	}

	// Clear existing items
	navContent.RemoveAll()
	sidebarItems = make([]*SidebarItem, 0, len(webapps))

	// Create new items
	for i, app := range webapps {
		appID := app.ID
		appIndex := i

		item := NewSidebarItem(app.Name, func() {
			utils.Dispatcher.Emit(utils.AppSelect, appID)
			updateSelection(appIndex)
		}, currentThemeVariant)

		sidebarItems = append(sidebarItems, item)
		navContent.Add(item)
	}

	navContent.Refresh()
}

func updateSelection(index int) {
	for i, item := range sidebarItems {
		item.SetSelected(i == index)
	}
	currentSelectedID = index
}
