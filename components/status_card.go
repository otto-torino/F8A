package components

import (
	"fmt"
	"image/color"
	"os/exec"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
)

type StatusCard struct {
	widget.BaseWidget
	app        *models.App
	statusText *canvas.Text
	remoteText *canvas.Text
	localText  *canvas.Text
	refreshBtn *widget.Button
}

func NewStatusCard(app *models.App) *StatusCard {
	statusText := canvas.NewText("Status: Checking...", color.RGBA{R: 200, G: 200, B: 200, A: 255})
	statusText.TextSize = 14
	statusText.TextStyle = fyne.TextStyle{Bold: true}

	remoteText := canvas.NewText("Remote: Checking...", color.RGBA{R: 200, G: 200, B: 200, A: 255})
	remoteText.TextSize = 12

	localText := canvas.NewText("Local: Checking...", color.RGBA{R: 200, G: 200, B: 200, A: 255})
	localText.TextSize = 12

	card := &StatusCard{
		app:        app,
		statusText: statusText,
		remoteText: remoteText,
		localText:  localText,
	}

	card.refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		go card.UpdateStatus()
	})

	card.ExtendBaseWidget(card)
	card.Refresh()
	return card
}

func (sc *StatusCard) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.RGBA{R: 40, G: 40, B: 40, A: 255})

	header := container.NewHBox(sc.statusText, layout.NewSpacer(), sc.refreshBtn)

	content := container.NewVBox(
		header,
		container.NewHBox(sc.localText, layout.NewSpacer()),
		container.NewHBox(sc.remoteText, layout.NewSpacer()),
	)

	padded := container.New(layout.NewPaddedLayout(), content)

	return &statusCardRenderer{
		background: background,
		content:    padded,
		objects:    []fyne.CanvasObject{background, padded},
	}
}

type statusCardRenderer struct {
	background *canvas.Rectangle
	content    *fyne.Container
	objects    []fyne.CanvasObject
}

func (r *statusCardRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *statusCardRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

func (r *statusCardRenderer) Refresh() {
	r.background.Refresh()
	r.content.Refresh()
}

func (r *statusCardRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *statusCardRenderer) Destroy() {}

func (sc *StatusCard) UpdateStatus() {
	// Get local revision
	localHash := sc.getLocalRevision()
	if localHash == "" {
		sc.statusText.Text = "Status: Error getting local revision"
		sc.statusText.Color = color.RGBA{R: 255, G: 100, B: 100, A: 255}
		sc.localText.Text = "Local: Error"
		sc.localText.Color = color.White
		sc.statusText.Refresh()
		sc.localText.Refresh()
		return
	}

	// Update local revision immediately
	sc.localText.Text = fmt.Sprintf("Local: %s", localHash[:7])
	sc.localText.Color = color.White
	sc.localText.Refresh()

	// Get remote revision from actual server
	remoteHash := sc.getRemoteRevision()
	if remoteHash == "" {
		sc.statusText.Text = "Status: Never deployed"
		sc.statusText.Color = color.RGBA{R: 255, G: 200, B: 100, A: 255}
		sc.remoteText.Text = "Remote: N/A"
		sc.remoteText.Color = color.White
		sc.statusText.Refresh()
		sc.remoteText.Refresh()
		return
	}

	// Update remote revision
	sc.remoteText.Text = fmt.Sprintf("Remote: %s", remoteHash)
	sc.remoteText.Color = color.White

	// Compare revisions
	if localHash[:7] == remoteHash {
		sc.statusText.Text = "Status: Up to date"
		sc.statusText.Color = color.RGBA{R: 100, G: 255, B: 100, A: 255}
	} else {
		// Check if local is ahead
		commitsAhead := sc.getCommitsAhead(remoteHash, localHash)
		if commitsAhead > 0 {
			sc.statusText.Text = fmt.Sprintf("Status: Local changes (+%d commits)", commitsAhead)
		} else {
			sc.statusText.Text = "Status: Local changes"
		}
		sc.statusText.Color = color.RGBA{R: 255, G: 200, B: 100, A: 255}
	}

	// Get deployment time from database if available
	lastDeployment, err := models.GetLastSuccessfulDeployment(sc.app.ID)
	if err == nil && lastDeployment != nil && lastDeployment.CommitHash == remoteHash {
		timeAgo := formatTimeAgo(lastDeployment.StartedAt)
		sc.remoteText.Text = fmt.Sprintf("Remote: %s (deployed %s)", remoteHash, timeAgo)
	}

	sc.statusText.Refresh()
	sc.remoteText.Refresh()
}

func (sc *StatusCard) getLocalRevision() string {
	out, err := exec.Command("bash", "-c", "cd "+sc.app.LocalPath+" && git rev-parse HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (sc *StatusCard) getRemoteRevision() string {
	cmd := fmt.Sprintf("ssh otto@%s readlink -f %s/%s", sc.app.RemoteHost, sc.app.RemotePath, sc.app.CurrentDirName)
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return ""
	}
	fullPath := strings.TrimSpace(string(out))
	// Extract revision from path (last part after /)
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func (sc *StatusCard) getCommitsAhead(baseHash, headHash string) int {
	out, err := exec.Command("bash", "-c",
		fmt.Sprintf("cd %s && git rev-list --count %s..%s", sc.app.LocalPath, baseHash, headHash),
	).Output()
	if err != nil {
		return 0
	}
	var count int
	fmt.Sscanf(string(out), "%d", &count)
	return count
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
