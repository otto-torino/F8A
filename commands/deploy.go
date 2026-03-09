package commands

import (
	"errors"
	"fmt"
	"image/color"
	"net/http"
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
	tracker := progress.NewDeploymentProgress(deploymentID, app.ID, commitHash, app.HealthCheckUrl)
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
			err = deployWithProgress(app, commitHash, tracker)
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

func deployWithProgress(app *models.App, commitHash string, tracker *progress.DeploymentProgress) error {
	// Check if already deployed
	if utils.CheckRemoteRevisionEqualsLocal(app) {
		return errors.New("Revision already deployed")
	}

	// Step 1: Build
	tracker.StartStep(progress.StepBuild)
	out, err := utils.ExecCommand(fmt.Sprintf("cd %s && yarn build", app.LocalPath))
	tracker.CompleteStep(progress.StepBuild, err == nil, out, err)
	if err != nil {
		return err
	}

	// Step 2: Archive
	tracker.StartStep(progress.StepArchive)
	out, err = utils.ExecCommand(fmt.Sprintf("cd %s && tar cvf %s.tar %s", app.LocalPath, commitHash, app.LocalDistDirName))
	tracker.CompleteStep(progress.StepArchive, err == nil, out, err)
	if err != nil {
		return err
	}

	// Step 3: Upload
	tracker.StartStep(progress.StepUpload)
	out, err = utils.ExecCommand(fmt.Sprintf("scp %s/%s.tar otto@%s:%s", app.LocalPath, commitHash, app.RemoteHost, app.RemotePath))
	tracker.CompleteStep(progress.StepUpload, err == nil, out, err)
	if err != nil {
		return err
	}

	// Skip ls command (not part of critical path)
	utils.ExecCommand(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath))

	// Step 4: Backup
	tracker.StartStep(progress.StepBackup)
	out1, err := utils.ExecCommand(fmt.Sprintf("ssh otto@%s rm -r %s/previous", app.RemoteHost, app.RemotePath))
	var out2 string
	if err == nil {
		out2, err = utils.ExecCommand(fmt.Sprintf("ssh otto@%s mv %s/%s %s/previous", app.RemoteHost, app.RemotePath, app.CurrentDirName, app.RemotePath))
	}
	tracker.CompleteStep(progress.StepBackup, err == nil, out1+"\n"+out2, err)
	if err != nil {
		return err
	}

	// Step 5: Extract
	tracker.StartStep(progress.StepExtract)
	out1, err = utils.ExecCommand(fmt.Sprintf("ssh otto@%s tar xvf %s/%s.tar -C %s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath))
	out2 = ""
	var out3 string
	if err == nil {
		// Ensure destination is clean before move
		out2, _ = utils.ExecCommand(fmt.Sprintf("ssh otto@%s rm -rf %s/%s", app.RemoteHost, app.RemotePath, commitHash))
		out3, err = utils.ExecCommand(fmt.Sprintf("ssh otto@%s mv %s/%s %s/%s", app.RemoteHost, app.RemotePath, app.LocalDistDirName, app.RemotePath, commitHash))
	}
	tracker.CompleteStep(progress.StepExtract, err == nil, out1+"\n"+out2+"\n"+out3, err)
	if err != nil {
		return err
	}

	// Skip ls command (not part of critical path)
	utils.ExecCommand(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath))

	// Step 6: Activate
	tracker.StartStep(progress.StepActivate)
	out, err = utils.ExecCommand(fmt.Sprintf("ssh otto@%s ln -s %s/%s %s/%s", app.RemoteHost, app.RemotePath, commitHash, app.RemotePath, app.CurrentDirName))
	tracker.CompleteStep(progress.StepActivate, err == nil, out, err)
	if err != nil {
		return err
	}

	// Step 7: Cleanup
	tracker.StartStep(progress.StepCleanup)
	out1, err = utils.ExecCommand(fmt.Sprintf("ssh otto@%s rm %s/%s.tar", app.RemoteHost, app.RemotePath, commitHash))
	out2 = ""
	if err == nil && app.HasHtAccess == 1 {
		out2, err = utils.ExecCommand(
			fmt.Sprintf("ssh otto@%s cp %s/.htaccess %s/%s", app.RemoteHost, app.RemotePath, app.RemotePath, app.CurrentDirName),
		)
	}
	tracker.CompleteStep(progress.StepCleanup, err == nil, out1+"\n"+out2, err)
	if err != nil {
		return err
	}

	// Step 8: Health check (only if URL is configured)
	if app.HealthCheckUrl != "" {
		tracker.StartStep(progress.StepHealthCheck)
		healthOut, healthErr := performHealthCheck(app.HealthCheckUrl)
		tracker.CompleteStep(progress.StepHealthCheck, healthErr == nil, healthOut, healthErr)
		if healthErr != nil {
			return healthErr
		}
	}

	return nil
}

func performHealthCheck(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Sprintf("Health check failed: %s", err.Error()), fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	result := fmt.Sprintf("GET %s -> %d %s", url, resp.StatusCode, resp.Status)
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return result, nil
	}
	return result, fmt.Errorf("health check failed with status %d", resp.StatusCode)
}

func Restore(app *models.App, onSuccess func()) func() {
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

				if err := utils.Shellout(RemovePrevSymlinkDesc, fmt.Sprintf("ssh otto@%s rm -r %s/previous", app.RemoteHost, app.RemotePath), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				if err := utils.Shellout(MoveCurrentRevToPreviousDesc, fmt.Sprintf("ssh otto@%s mv %s/%s %s/previous", app.RemoteHost, app.RemotePath, app.CurrentDirName, app.RemotePath), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				if err := utils.Shellout(ActivateCurrentSymlinkDesc, fmt.Sprintf("ssh otto@%s ln -s %s/%s %s/%s", app.RemoteHost, app.RemotePath, revision, app.RemotePath, app.CurrentDirName), outInfo, false); err != nil {
					utils.AddTextToOutput(err.Error(), errorColor, outInfo)
					return
				}
				utils.AddTextToOutput(fmt.Sprintf("Restored revision %s", revision), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outInfo)
				if onSuccess != nil {
					onSuccess()
				}
			})
			content.Add(btn)
		}
	}
}
