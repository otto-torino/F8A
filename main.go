package main

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/otto-torino/f8a/components"
	"github.com/otto-torino/f8a/db"
	"github.com/otto-torino/f8a/logger"
	"github.com/otto-torino/f8a/theme"
	"github.com/spf13/viper"
)

func init() {
	// Read settings
	viper.SetConfigFile(fmt.Sprintf("/etc/f8a.json"))
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Error reading settings file, %s", err))
	}

	// Ensure app home directory exists
	homePath := viper.GetString("app.homePath")

	err := os.MkdirAll(homePath, os.ModePerm)
	if err != nil {
		panic("Error creating the application home directory")
	}

	// Init logger
	logger.InitLogger()
	logger.ZapLog.Info("Starting F8A application")
}

func main() {
	// Init database
	db.InitDatabase()

	a := app.NewWithID("io.otto-torino.f8a")
	w := a.NewWindow("Otto Frontend Apps Manager")

	a.Settings().SetTheme(&theme.F8aTheme{})
	w.SetMainMenu(components.MakeMenu(a, w))

	mainContent := components.MakeMainContent()
	sidebar := components.MakeSidebar(components.HandleAddWebApp)
	mainLayout := container.NewHSplit(sidebar, mainContent)
	mainLayout.Offset = 0.3
	w.SetContent(mainLayout)

	w.Resize(fyne.NewSize(640, 460))
	w.ShowAndRun()
}
