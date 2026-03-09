package commands

import (
	"fmt"
	"image/color"
	"os/exec"
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

type DiffInfo struct {
	LocalHash     string
	RemoteHash    string
	CommitCount   int
	Commits       string
	FileSummary   string
	IsUpToDate    bool
	NeverDeployed bool
	BehindRemote  bool
}

func GetDiffPreview(app *models.App) *DiffInfo {
	info := &DiffInfo{}

	// Get local short hash
	out, err := exec.Command("bash", "-c", fmt.Sprintf("cd %s && git rev-parse --short HEAD", app.LocalPath)).Output()
	if err != nil {
		return nil
	}
	info.LocalHash = strings.TrimSpace(string(out))

	// Get remote revision
	remoteOut, err := exec.Command("bash", "-c",
		fmt.Sprintf("ssh otto@%s readlink -f %s/%s", app.RemoteHost, app.RemotePath, app.CurrentDirName)).Output()
	if err != nil {
		info.NeverDeployed = true
		// Get recent commits for never-deployed apps
		out, _ = exec.Command("bash", "-c",
			fmt.Sprintf("cd %s && git log --oneline -10", app.LocalPath)).Output()
		info.Commits = strings.TrimSpace(string(out))
		return info
	}

	fullPath := strings.TrimSpace(string(remoteOut))
	parts := strings.Split(fullPath, "/")
	remoteShortHash := parts[len(parts)-1]
	info.RemoteHash = remoteShortHash

	// Check if up to date
	if info.LocalHash == info.RemoteHash {
		info.IsUpToDate = true
		return info
	}

	// List recent commits and find where the remote hash appears
	// Using a generous limit to find the deployed commit in history
	out, _ = exec.Command("bash", "-c",
		fmt.Sprintf("cd %s && git log --oneline -200", app.LocalPath)).Output()
	allCommits := strings.TrimSpace(string(out))

	if allCommits != "" {
		lines := strings.Split(allCommits, "\n")
		var commitLines []string
		found := false
		for _, line := range lines {
			// Each line starts with a short hash; stop when we find the remote hash
			if strings.HasPrefix(line, remoteShortHash) {
				found = true
				break
			}
			commitLines = append(commitLines, line)
		}
		if found {
			info.Commits = strings.Join(commitLines, "\n")
			info.CommitCount = len(commitLines)
		} else {
			// Remote hash not found in local history — local is likely behind
			info.BehindRemote = true
		}
	}

	// Get diffstat using git diff with merge-base to handle short hash issues
	// First find the full hash by searching git log
	out, _ = exec.Command("bash", "-c",
		fmt.Sprintf("cd %s && git log --all --format='%%H' | grep ^%s | head -1", app.LocalPath, remoteShortHash)).Output()
	fullHash := strings.TrimSpace(string(out))
	if fullHash != "" {
		out, _ = exec.Command("bash", "-c",
			fmt.Sprintf("cd %s && git diff --stat %s..HEAD", app.LocalPath, fullHash)).Output()
		info.FileSummary = strings.TrimSpace(string(out))
	}

	return info
}

func ShowDiffPreview(app *models.App, onConfirm func()) {
	info := GetDiffPreview(app)
	if info == nil {
		return
	}

	registry := utils.Registry()

	if info.IsUpToDate {
		dialog.ShowInformation("Deploy Preview", "Already up to date — nothing to deploy.", *registry.Window)
		return
	}

	if info.BehindRemote {
		dialog.ShowConfirm(
			"Warning: Local may be behind remote",
			fmt.Sprintf(
				"The deployed revision (%s) was not found in your local git history.\n\n"+
					"Your local branch is likely behind the remote. You should pull before deploying.\n\n"+
					"Deploy anyway?",
				info.RemoteHash,
			),
			func(confirmed bool) {
				if confirmed {
					onConfirm()
				}
			},
			*registry.Window,
		)
		return
	}

	// Build the dialog content
	contentItems := container.NewVBox()

	// Header
	if info.NeverDeployed {
		header := canvas.NewText("First deployment", color.RGBA{R: 255, G: 200, B: 100, A: 255})
		header.TextSize = 14
		header.TextStyle = fyne.TextStyle{Bold: true}
		contentItems.Add(header)
		localLabel := canvas.NewText(fmt.Sprintf("Local: %s", info.LocalHash), color.White)
		localLabel.TextSize = 12
		contentItems.Add(localLabel)
	} else {
		header := canvas.NewText(
			fmt.Sprintf("%d commit(s) to deploy", info.CommitCount),
			color.RGBA{R: 255, G: 200, B: 100, A: 255},
		)
		header.TextSize = 14
		header.TextStyle = fyne.TextStyle{Bold: true}
		contentItems.Add(header)

		revInfo := canvas.NewText(
			fmt.Sprintf("%s (remote) -> %s (local)", info.RemoteHash, info.LocalHash),
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
		)
		revInfo.TextSize = 12
		contentItems.Add(revInfo)
	}

	contentItems.Add(widget.NewSeparator())

	// Commits section
	commitsLabel := canvas.NewText("Commits:", color.White)
	commitsLabel.TextSize = 13
	commitsLabel.TextStyle = fyne.TextStyle{Bold: true}
	contentItems.Add(commitsLabel)

	if info.Commits != "" {
		commitsEntry := widget.NewMultiLineEntry()
		commitsEntry.SetText(info.Commits)
		commitsEntry.Wrapping = fyne.TextWrapWord
		commitsEntry.SetMinRowsVisible(min(strings.Count(info.Commits, "\n")+1, 10))
		contentItems.Add(commitsEntry)
	} else {
		contentItems.Add(widget.NewLabel("No commits found"))
	}

	// File changes section
	if info.FileSummary != "" {
		contentItems.Add(widget.NewSeparator())

		filesLabel := canvas.NewText("Changed files:", color.White)
		filesLabel.TextSize = 13
		filesLabel.TextStyle = fyne.TextStyle{Bold: true}
		contentItems.Add(filesLabel)

		filesEntry := widget.NewMultiLineEntry()
		filesEntry.SetText(info.FileSummary)
		filesEntry.Wrapping = fyne.TextWrapWord
		filesEntry.SetMinRowsVisible(min(strings.Count(info.FileSummary, "\n")+1, 10))
		contentItems.Add(filesEntry)
	}

	scroll := container.NewScroll(container.New(layout.NewPaddedLayout(), contentItems))

	d := dialog.NewCustomConfirm("Deploy Preview", "Deploy", "Cancel", scroll, func(confirmed bool) {
		if confirmed {
			onConfirm()
		}
	}, *registry.Window)
	d.Resize(fyne.NewSize(600, 500))
	d.Show()
}
