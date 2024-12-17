package commands

import (
	"errors"
	"fmt"
	"image/color"
	"os/exec"

	"fyne.io/fyne/v2"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)

func Deploy(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		outputContainer.RemoveAll()
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
	}
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
