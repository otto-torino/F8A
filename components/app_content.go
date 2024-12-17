package components

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/commands"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
	"go.uber.org/zap"
)

var errorColor = color.RGBA{255, 0, 0, 255}

func HandleAddWebApp() {
	mainContent.RemoveAll()

	errorText := canvas.NewText("", errorColor)

	hasHtAccess := false
	labelName := widget.NewLabel("App Name")
	name := widget.NewEntry()
	labelPath := widget.NewLabel("App Local Path")
	localPath := widget.NewEntry()
	labelRemoteHost := widget.NewLabel("App Remote Host")
	remoteHost := widget.NewEntry()
	labelRemotePath := widget.NewLabel("App Remote Path")
	remotePath := widget.NewEntry()
	labelCurrentDirName := widget.NewLabel("Current Directory Name")
	currentDirName := widget.NewEntry()
	currentDirName.SetText("current")
	hasHtAccessLabel := widget.NewLabel("Requires .htaccess")
	hasHtAccessWidget := widget.NewCheck("", func(b bool) {
		hasHtAccess = b
		fmt.Println(hasHtAccess)
	})
	grid := container.New(layout.NewFormLayout(), labelName, name, labelPath, localPath, labelRemoteHost, remoteHost, labelRemotePath, remotePath, labelCurrentDirName, currentDirName, hasHtAccessLabel, hasHtAccessWidget)

	addButton := widget.NewButton("Save", func() {
		if name.Text == "" || localPath.Text == "" || remotePath.Text == "" || remoteHost.Text == "" {
			errorText.Text = "Please fill all fields"
			errorText.Refresh()
			return
		}
		id, err := models.CreateApp(name.Text, localPath.Text, remoteHost.Text, remotePath.Text, currentDirName.Text, hasHtAccess)
		if err != nil {
			errorText.Text = err.Error()
			errorText.Refresh()
			return
		}
		utils.Dispatcher.Emit(utils.AppChange)
		utils.Dispatcher.Emit(utils.AppAdd, int(id))
	})

	mainContent.Add(container.New(layout.NewVBoxLayout(), grid, errorText, addButton))
}

func HandleChangeWebApp(id int) {
	app, err := models.GetApp(id)
	if err != nil {
		zap.S().Error("Cannot get app", err)
		return
	}

	mainContent.RemoveAll()

	errorText := canvas.NewText("", color.RGBA{255, 0, 0, 255})

	hasHtAccess := app.HasHtAccess != 0
	labelName := widget.NewLabel("App Name")
	name := widget.NewEntry()
	name.SetText(app.Name)
	labelPath := widget.NewLabel("App Local Path")
	localPath := widget.NewEntry()
	localPath.SetText(app.LocalPath)
	labelRemoteHost := widget.NewLabel("App Remote Host")
	remoteHost := widget.NewEntry()
	remoteHost.SetText(app.RemoteHost)
	labelRemotePath := widget.NewLabel("App Remote Path")
	remotePath := widget.NewEntry()
	remotePath.SetText(app.RemotePath)
	labelCurrentDirName := widget.NewLabel("Current Directory Name")
	currentDirName := widget.NewEntry()
	currentDirName.SetText(app.CurrentDirName)
	hasHtAccessLabel := widget.NewLabel("Requires .htaccess")
	hasHtAccessWidget := widget.NewCheck("", func(b bool) {
		hasHtAccess = b
		fmt.Println(hasHtAccess)
	})
	hasHtAccessWidget.SetChecked(hasHtAccess)
	grid := container.New(layout.NewFormLayout(), labelName, name, labelPath, localPath, labelRemoteHost, remoteHost, labelRemotePath, remotePath, labelCurrentDirName, currentDirName, hasHtAccessLabel, hasHtAccessWidget)

	changeButton := widget.NewButton("Save", func() {
		if name.Text == "" || localPath.Text == "" || remotePath.Text == "" || remoteHost.Text == "" {
			errorText.Text = "Please fill all fields"
			errorText.Refresh()
			return
		}
		err := models.UpdateApp(id, name.Text, localPath.Text, remoteHost.Text, remotePath.Text, currentDirName.Text, hasHtAccess)
		if err != nil {
			errorText.Text = err.Error()
			errorText.Refresh()
			return
		}
		utils.Dispatcher.Emit(utils.AppChange)
		utils.Dispatcher.Emit(utils.AppUpdate)
	})

	mainContent.Add(container.New(layout.NewVBoxLayout(), grid, errorText, changeButton))
}

func HandleWebAppSection(id int) {
	app, err := models.GetApp(id)
	if err != nil {
		zap.S().Error("Cannot get app", err)
		return
	}

	// clean
	mainContent.RemoveAll()

	// top title and delete button
	title := utils.MakeTitle(app.Name)
	editButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		HandleChangeWebApp(id)
	})
	deleteButton := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		_ = app.Delete()
		utils.Dispatcher.Emit(utils.AppDelete)
	})
	header := container.NewHBox(title, layout.NewSpacer(), editButton, deleteButton)

	nameLabel := widget.NewLabel("App Name")
	name := widget.NewLabel(app.Name)
	localPathLabel := widget.NewLabel("App Local Path")
	localPath := widget.NewLabel(app.LocalPath)
	remoteHostLabel := widget.NewLabel("App Remote Host")
	remoteHost := widget.NewLabel(app.RemoteHost)
	remotePathLabel := widget.NewLabel("App Remote Path")
	remotePath := widget.NewLabel(app.RemotePath)
	currentDirNameLabel := widget.NewLabel("Current Directory Name")
	currentDirName := widget.NewLabel(app.CurrentDirName)
	hasHtAccessLabel := widget.NewLabel("Requires .htaccess")
	hasHtAccessStr := "no"
	if app.HasHtAccess != 0 {
		hasHtAccessStr = "yes"
	}
	hasHtAccess := widget.NewLabel(hasHtAccessStr)
	infoGrid := container.New(layout.NewFormLayout(), nameLabel, name, localPathLabel, localPath, remoteHostLabel, remoteHost, remotePathLabel, remotePath, currentDirNameLabel, currentDirName, hasHtAccessLabel, hasHtAccess)

	top := container.NewVBox(header, infoGrid)

	output := container.NewVBox()
	background := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	utils.Scroll = container.NewScroll(output)
	outputContainer := container.NewBorder(nil, nil, nil, nil, container.NewStack(background, utils.Scroll))

	actionButtons := MakeActionButtons(app, output)

	mainContent.Add(container.NewBorder(top, actionButtons, nil, nil, outputContainer))
}

func MakeActionButtons(app *models.App, outputContainer *fyne.Container) *fyne.Container {

	// Build Button
	build := utils.MakeButton("Build", commands.Build(app, outputContainer))

	// Build Archive
	buildArchive := utils.MakeButton("Build Archive", commands.BuildArchive(app, outputContainer))

	// Local revision button
	gitLocaleRev := utils.MakeButton("Local Revision", commands.LocalRevision(app, outputContainer))

	// Remote revision button
	gitRemoteRev := utils.MakeButton("Remote Revision", commands.RemoteRevision(app, outputContainer))

	// Deploy button
	deploy := utils.MakeButton("Deploy", commands.Deploy(app, outputContainer))

	// Restore button
	restore := utils.MakeButton("Restore Revision", func() {
		return
	})

	return container.NewHBox(build, buildArchive, gitLocaleRev, gitRemoteRev, deploy, restore)
}
