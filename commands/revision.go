package commands

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)

func LocalRevision(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		outputContainer.RemoveAll()
		out, err := utils.Shell(fmt.Sprintf("cd %s && git rev-parse --short HEAD", app.LocalPath))
		if err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
		utils.AddTextToOutput(fmt.Sprintf("Git Local Revision"), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outputContainer)
		utils.AddTextToOutput(fmt.Sprintf("%s", (*out)[0]), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
		return
	}
}

func RemoteRevision(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		outputContainer.RemoveAll()
		utils.AddTextToOutput(fmt.Sprintf("Retrieving Git Remote Revision..."), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outputContainer)
		out, err := utils.Shell(fmt.Sprintf("ssh otto@%s readlink -f %s/%s", app.RemoteHost, app.RemotePath, app.CurrentDirName))
		outputContainer.RemoveAll()
		if err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
		utils.AddTextToOutput(fmt.Sprintf("Git Remote Revision"), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outputContainer)
		o := (*out)[0]
		rev := o[strings.LastIndex(o, "/")+1:]
		utils.AddTextToOutput(fmt.Sprintf("%s", rev), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
		return
	}
}
