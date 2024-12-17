package commands

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)

var errorColor = color.RGBA{255, 0, 0, 255}

func Build(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		if err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
	}
}

func BuildArchive(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		if err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
		if err := utils.Shellout(fmt.Sprintf("cd %s && tar cvf dist.tar dist", app.LocalPath), outputContainer, false); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
	}
}
