package components

import (
	"errors"
	"fmt"
	"image/color"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
	build := utils.MakeButton("Build", func() {
		if err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
	})

	// Build Archive
	buildArchive := utils.MakeButton("Build Archive", func() {
		if err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
		if err := utils.Shellout(fmt.Sprintf("cd %s && tar cvf dist.tar dist", app.LocalPath), outputContainer, false); err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
	})

	// Local revision button
	gitLocaleRev := utils.MakeButton("Git Local Revision", func() {
		out, err := utils.Shell(fmt.Sprintf("cd %s && git rev-parse --short HEAD", app.LocalPath))
		if err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outputContainer)
			return
		}
		utils.AddTextToOutput(fmt.Sprintf("Git Local Revision"), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outputContainer)
		utils.AddTextToOutput(fmt.Sprintf("%s", (*out)[0]), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
		return
	})

	// Remote revision button
	gitRemoteRev := utils.MakeButton("Git Remote Revision", func() {
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
	})

	// Deploy button
	deploy := utils.MakeButton("Deploy", func() {
		out, err := exec.Command("bash", "-c", "cd "+app.LocalPath+" && git rev-parse --short HEAD").Output()
		if err != nil {
			fmt.Println(err)
		}
		commitHash := string(out)[0:7]

		utils.AddTextToOutput("Deploying revision "+commitHash, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)
		err = deploy(app, outputContainer, commitHash)
		if err != nil {
			utils.AddTextToOutput("Deployment failed for revision "+commitHash, errorColor, outputContainer)
			return
		}
		utils.AddTextToOutput("Deployed revision "+commitHash, color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
	})

	return container.NewHBox(build, buildArchive, gitLocaleRev, gitRemoteRev, deploy)
}

func deploy(app *models.App, outputContainer *fyne.Container, commitHash string) error {
	if utils.CheckRemoteRevisionEqualsLocal(app) {
		utils.AddTextToOutput("Revision already deployed", errorColor, outputContainer)
		return errors.New("Revision already deployed")
	}
	if err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("cd %s && tar cvf %s.tar dist", app.LocalPath, commitHash), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("scp %s/%s.tar otto@%s:%s", app.LocalPath, commitHash, app.RemoteHost, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s rm -r %s/previous", app.RemoteHost, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/%s %s/previous", app.RemoteHost, app.RemotePath, app.CurrentDirName, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s tar xvf %s/%s.tar -C %s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/dist %s/%s", app.RemoteHost, app.RemotePath, app.RemotePath, commitHash), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s ln -s %s/%s %s/%s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath, app.CurrentDirName), outputContainer, false); err != nil {
		return err
	}
	if err := utils.Shellout(fmt.Sprintf("ssh otto@%s rm %s/%s.tar", app.RemoteHost, app.RemotePath, commitHash), outputContainer, false); err != nil {
		return err
	}
	if app.HasHtAccess == 1 {
		if err := utils.Shellout(
			fmt.Sprintf("ssh otto@%s cp %s/.htaccess %s/%s", app.RemoteHost, app.RemotePath, app.RemotePath, app.CurrentDirName),
			outputContainer,
			false,
		); err != nil {
			return err
		}
	}

	return nil
}
