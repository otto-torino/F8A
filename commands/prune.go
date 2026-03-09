package commands

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)

type remoteDeployment struct {
	name    string
	size    string
	modTime string
}

func Prune(app *models.App, onSuccess func()) func() {
	return func() {
		background := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
		background.SetMinSize(fyne.NewSize(500, 80))
		outInfo := container.NewVBox()
		utils.Scroll = container.NewScroll(container.New(layout.NewPaddedLayout(), outInfo))

		content := container.NewVBox()
		contentScroll := container.NewScroll(content)

		c := container.NewBorder(container.NewStack(background, utils.Scroll), nil, nil, nil, contentScroll)

		registry := utils.Registry()
		d := dialog.NewCustom("Prune old deployments", "Dismiss", c, *registry.Window)
		d.Show()
		d.Resize(fyne.NewSize(500, 500))

		utils.AddTextToOutput("Retrieving remote deployments...", color.RGBA{R: 255, G: 255, B: 255, A: 255}, outInfo)

		// Get current symlink target to exclude it
		currentOut, err := utils.Shell(fmt.Sprintf("ssh otto@%s readlink -f %s/%s", app.RemoteHost, app.RemotePath, app.CurrentDirName))
		if err != nil {
			utils.AddTextToOutput("Failed to read current symlink: "+err.Error(), errorColor, outInfo)
			return
		}
		currentTarget := ""
		if currentOut != nil && len(*currentOut) > 0 {
			parts := strings.Split((*currentOut)[0], "/")
			currentTarget = parts[len(parts)-1]
		}

		// List all directories with sizes
		sizeOut, err := utils.Shell(fmt.Sprintf(
			"ssh otto@%s du -sh %s/*/",
			app.RemoteHost, app.RemotePath,
		))
		if err != nil {
			utils.AddTextToOutput("Failed to list deployments: "+err.Error(), errorColor, outInfo)
			return
		}

		if sizeOut == nil || len(*sizeOut) == 0 {
			utils.AddTextToOutput("No old deployments found.", color.RGBA{R: 100, G: 255, B: 100, A: 255}, outInfo)
			return
		}

		// Get modification times using ls --full-time
		// Output format: drwxr-xr-x 2 otto otto 4096 2026-03-05 14:30:00.000000000 +0100 dirname
		timeRaw, _ := utils.ExecCommand(fmt.Sprintf(
			"ssh otto@%s ls -ld --full-time %s/*/",
			app.RemoteHost, app.RemotePath,
		))
		modTimes := make(map[string]string)
		for _, line := range strings.Split(timeRaw, "\n") {
			fields := strings.Fields(line)
			// Expect at least 9 fields: perms, links, owner, group, size, date, time, tz, name
			if len(fields) < 9 {
				continue
			}
			dateStr := fields[5]
			timeStr := fields[6]
			// Trim sub-second precision for display
			if idx := strings.Index(timeStr, "."); idx != -1 {
				timeStr = timeStr[:idx]
			}
			dirPath := strings.TrimSuffix(fields[len(fields)-1], "/")
			parts := strings.Split(dirPath, "/")
			dirName := parts[len(parts)-1]
			modTimes[dirName] = dateStr + " " + timeStr
		}

		// Parse the output into deployments
		var deployments []remoteDeployment
		for _, line := range *sizeOut {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			size := fields[0]
			dirPath := strings.TrimSuffix(fields[1], "/")
			parts := strings.Split(dirPath, "/")
			name := parts[len(parts)-1]

			// Skip current symlink, previous symlink, and the active deployment directory
			if name == app.CurrentDirName || name == "previous" || name == currentTarget {
				continue
			}

			mt := modTimes[name]
			if mt == "" {
				mt = "unknown"
			}
			deployments = append(deployments, remoteDeployment{name: name, size: size, modTime: mt})
		}

		if len(deployments) == 0 {
			utils.AddTextToOutput("No old deployments to prune.", color.RGBA{R: 100, G: 255, B: 100, A: 255}, outInfo)
			return
		}

		utils.AddTextToOutput(fmt.Sprintf("Found %d old deployment(s). Select the ones to delete:", len(deployments)), color.RGBA{R: 255, G: 255, B: 255, A: 255}, outInfo)

		// Build checkboxes for each deployment
		selected := make(map[string]bool)
		checks := make([]*widget.Check, 0, len(deployments))

		for _, dep := range deployments {
			dep := dep
			label := fmt.Sprintf("%s  -  %s  (%s)", dep.name, dep.modTime, dep.size)
			check := widget.NewCheck(label, func(checked bool) {
				selected[dep.name] = checked
			})
			checks = append(checks, check)
			content.Add(check)
		}

		// Select All / Deselect All
		selectAll := widget.NewCheck("Select All", func(checked bool) {
			for i, dep := range deployments {
				selected[dep.name] = checked
				checks[i].SetChecked(checked)
			}
		})
		content.Add(widget.NewSeparator())
		content.Add(selectAll)

		// Delete button
		deleteBtn := widget.NewButton("Delete Selected", func() {
			toDelete := []string{}
			for name, sel := range selected {
				if sel {
					toDelete = append(toDelete, name)
				}
			}
			if len(toDelete) == 0 {
				utils.AddTextToOutput("No deployments selected.", color.RGBA{R: 255, G: 200, B: 100, A: 255}, outInfo)
				return
			}

			// Build rm commands for selected directories
			rmParts := make([]string, len(toDelete))
			for i, name := range toDelete {
				rmParts[i] = fmt.Sprintf("%s/%s", app.RemotePath, name)
			}
			rmCmd := fmt.Sprintf("ssh otto@%s rm -rf %s", app.RemoteHost, strings.Join(rmParts, " "))

			utils.AddTextToOutput(fmt.Sprintf("Deleting %d deployment(s)...", len(toDelete)), color.RGBA{R: 255, G: 153, B: 0, A: 255}, outInfo)

			_, err := utils.ExecCommand(rmCmd)
			if err != nil {
				utils.AddTextToOutput("Failed to delete: "+err.Error(), errorColor, outInfo)
				return
			}

			utils.AddTextToOutput(fmt.Sprintf("Successfully deleted %d deployment(s).", len(toDelete)), color.RGBA{R: 0, G: 255, B: 0, A: 255}, outInfo)

			// Remove deleted items from the UI
			content.RemoveAll()
			content.Add(widget.NewLabel("Pruning complete."))
			content.Refresh()

			if onSuccess != nil {
				onSuccess()
			}
		})
		deleteBtn.Importance = widget.DangerImportance
		content.Add(widget.NewSeparator())
		content.Add(deleteBtn)
	}
}
