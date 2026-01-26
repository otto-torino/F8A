package commands

import (
	"errors"
	"fmt"
	"image/color"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/progress"
	"github.com/otto-torino/f8a/utils"
)

func Deploy(app *models.App, outputContainer *fyne.Container, onComplete func()) (*progress.DeploymentProgress, func()) {
	outputContainer.RemoveAll()

	// Get commit hash
	out, err := exec.Command("bash", "-c", "cd "+app.LocalPath+" && git rev-parse --short HEAD").Output()
	if err != nil {
		utils.AddTextToOutput("Failed to get commit hash: "+err.Error(), errorColor, outputContainer)
		return nil, nil
	}
	commitHash := strings.TrimSpace(string(out))

	// Create deployment record
	deploymentID, err := models.CreateDeployment(app.ID, commitHash)
	if err != nil {
		utils.AddTextToOutput("Failed to create deployment record: "+err.Error(), errorColor, outputContainer)
		return nil, nil
	}

	// Initialize progress tracker
	tracker := progress.NewDeploymentProgress(deploymentID, app.ID, commitHash)
	tracker.LoadAverageDurations()

	return tracker, func() {
		go func() {
			defer tracker.Close()
			if onComplete != nil {
				defer onComplete()
			}

			utils.AddTextToOutput("Deploying revision "+commitHash, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)

			// Start deployment
			startTime := time.Now()
			err = deployWithProgress(app, outputContainer, commitHash, tracker)
			duration := time.Since(startTime)

			if err != nil {
				utils.AddTextToOutput("Deployment failed for revision "+commitHash, errorColor, outputContainer)
				errMsg := err.Error()
				currentStep := tracker.CurrentStep
				models.UpdateDeploymentStatus(deploymentID, "failed", &errMsg, &currentStep)
				utils.SendFailureNotification(app.Name, tracker.CurrentStep)
				return
			}

			utils.AddTextToOutput("Deployed revision "+commitHash, color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
			models.UpdateDeploymentStatus(deploymentID, "success", nil, nil)
			durationMs := duration.Milliseconds()
			models.UpdateDeploymentDuration(deploymentID, &durationMs)
			utils.SendSuccessNotification(app.Name, commitHash, duration)
		}()
	}
}

func deployWithProgress(app *models.App, outputContainer *fyne.Container, commitHash string, tracker *progress.DeploymentProgress) error {
	// Check if already deployed
	if utils.CheckRemoteRevisionEqualsLocal(app) {
		utils.AddTextToOutput("Revision already deployed", errorColor, outputContainer)
		return errors.New("Revision already deployed")
	}

	// Step 1: Build
	tracker.StartStep(progress.StepBuild)
	err := utils.Shellout(fmt.Sprintf("cd %s && yarn build", app.LocalPath), outputContainer, true)
	tracker.CompleteStep(progress.StepBuild, err == nil, "", err)
	if err != nil {
		return err
	}

	// Step 2: Archive
	tracker.StartStep(progress.StepArchive)
	err = utils.Shellout(fmt.Sprintf("cd %s && tar cvf %s.tar %s", app.LocalPath, commitHash, app.LocalDistDirName), outputContainer, false)
	tracker.CompleteStep(progress.StepArchive, err == nil, "", err)
	if err != nil {
		return err
	}

	// Step 3: Upload
	tracker.StartStep(progress.StepUpload)
	err = utils.Shellout(fmt.Sprintf("scp %s/%s.tar otto@%s:%s", app.LocalPath, commitHash, app.RemoteHost, app.RemotePath), outputContainer, false)
	tracker.CompleteStep(progress.StepUpload, err == nil, "", err)
	if err != nil {
		return err
	}

	// Skip ls command (not part of critical path)
	utils.Shellout(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath), outputContainer, false)

	// Step 4: Backup
	tracker.StartStep(progress.StepBackup)
	err = utils.Shellout(fmt.Sprintf("ssh otto@%s rm -r %s/previous", app.RemoteHost, app.RemotePath), outputContainer, false)
	if err == nil {
		err = utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/%s %s/previous", app.RemoteHost, app.RemotePath, app.CurrentDirName, app.RemotePath), outputContainer, false)
	}
	tracker.CompleteStep(progress.StepBackup, err == nil, "", err)
	if err != nil {
		return err
	}

	// Step 5: Extract
	tracker.StartStep(progress.StepExtract)
	err = utils.Shellout(fmt.Sprintf("ssh otto@%s tar xvf %s/%s.tar -C %s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath), outputContainer, false)
	if err == nil {
		// Ensure destination is clean before move
		utils.Shellout(fmt.Sprintf("ssh otto@%s rm -rf %s/%s", app.RemoteHost, app.RemotePath, commitHash), outputContainer, false)
		err = utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/%s %s/%s", app.RemoteHost, app.RemotePath, app.LocalDistDirName, app.RemotePath, commitHash), outputContainer, false)
	}
	tracker.CompleteStep(progress.StepExtract, err == nil, "", err)
	if err != nil {
		return err
	}

	// Skip ls command (not part of critical path)
	err = utils.Shellout(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath), outputContainer, false)
	if err != nil {
		return err
	}

	// Step 6: Activate
	tracker.StartStep(progress.StepActivate)
	err = utils.Shellout(fmt.Sprintf("ssh otto@%s ln -s %s/%s %s/%s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath, app.CurrentDirName), outputContainer, false)
	tracker.CompleteStep(progress.StepActivate, err == nil, "", err)
	if err != nil {
		return err
	}

	// Step 7: Cleanup
	tracker.StartStep(progress.StepCleanup)
	err = utils.Shellout(fmt.Sprintf("ssh otto@%s rm %s/%s.tar", app.RemoteHost, app.RemotePath, commitHash), outputContainer, false)
	if err == nil && app.HasHtAccess == 1 {
		err = utils.Shellout(
			fmt.Sprintf("ssh otto@%s cp %s/.htaccess %s/%s", app.RemoteHost, app.RemotePath, app.RemotePath, app.CurrentDirName),
			outputContainer,
			false,
		)
	}
	tracker.CompleteStep(progress.StepCleanup, err == nil, "", err)
	if err != nil {
		return err
	}

	return nil
}

func Restore(app *models.App) func() {
	return func() {
		background := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
		background.SetMinSize(fyne.NewSize(400, 80))
		outInfo := container.NewVBox()
		utils.Scroll = container.NewScroll(container.New(layout.NewPaddedLayout(), outInfo))

		content := container.NewVBox()
		c := container.NewBorder(container.NewStack(background, utils.Scroll), nil, nil, nil, container.NewScroll(content))

		registry := utils.Registry()
		d := dialog.NewCustom("Restore revision", "Dismiss", c, *registry.Window)
		d.Show()
		d.Resize(fyne.NewSize(400, 400))

		utils.AddTextToOutput(fmt.Sprintf("Retrieving available revisions..."), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outInfo)
		out, err := utils.Shell(fmt.Sprintf("ssh otto@%s ls -lst %s | grep -v previous | grep -v %s | tail -n +2 | awk '{print $7,$8,$9\" |\",$10}'", app.RemoteHost, app.RemotePath, app.CurrentDirName))
		if err != nil {
			utils.AddTextToOutput(err.Error(), errorColor, outInfo)
			return
		}

		utils.AddTextToOutput(fmt.Sprintf("Please click a revision to restore it"), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outInfo)

		for _, o := range *out {
			line := o
			btn := utils.MakeButton(o, func() {
				revision := line[strings.LastIndex(line, "|")+1:]
				revision = strings.TrimSpace(revision)

				utils.AddTextToOutput(fmt.Sprintf("Restoring revision %s", revision), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outInfo)

				if err := utils.Shellout(fmt.Sprintf("ssh otto@%s rm -r %s/previous", app.RemoteHost, app.RemotePath), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				if err := utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/%s %s/previous", app.RemoteHost, app.RemotePath, app.CurrentDirName, app.RemotePath), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				if err := utils.Shellout(fmt.Sprintf("ssh otto@%s ln -s %s/%s %s/%s", app.RemoteHost, app.RemotePath, revision, app.RemotePath, app.CurrentDirName), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				utils.AddTextToOutput(fmt.Sprintf("Restored revision %s", revision), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outInfo)
			})
			content.Add(btn)
		}
	}
}
