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
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
)

type StatusCard struct {
	widget.BaseWidget
	app          *models.App
	statusLabel  *widget.Label
	remoteLabel  *widget.Label
	localLabel   *widget.Label
	actionButton *widget.Button
}

func NewStatusCard(app *models.App) *StatusCard {
	card := &StatusCard{
		app:          app,
		statusLabel:  widget.NewLabel("Status: Loading..."),
		remoteLabel:  widget.NewLabel("Remote: ..."),
		localLabel:   widget.NewLabel("Local: ..."),
		actionButton: widget.NewButton("Deploy", func() {}),
	}
	card.ExtendBaseWidget(card)
	card.Refresh()
	return card
}

func (sc *StatusCard) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(color.RGBA{R: 240, G: 240, B: 240, A: 255})

	content := container.NewVBox(
		sc.statusLabel,
		sc.remoteLabel,
		sc.localLabel,
	)

	mainContent := container.NewBorder(
		nil, nil, nil, sc.actionButton,
		content,
	)

	padded := container.New(layout.NewPaddedLayout(), mainContent)

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
		sc.statusLabel.SetText("Status: Error getting local revision")
		return
	}

	// Get last successful deployment
	lastDeployment, err := models.GetLastSuccessfulDeployment(sc.app.ID)
	if err != nil || lastDeployment == nil {
		sc.statusLabel.SetText("Status: 🟡 Never deployed")
		sc.localLabel.SetText(fmt.Sprintf("Local: %s", localHash[:7]))
		sc.remoteLabel.SetText("Remote: N/A")
		sc.actionButton.SetText("Deploy")
		return
	}

	remoteHash := lastDeployment.CommitHash

	// Compare revisions
	if localHash[:7] == remoteHash {
		sc.statusLabel.SetText("Status: 🟢 Up to date")
		sc.actionButton.SetText("Re-deploy")
	} else {
		// Check if local is ahead
		commitsAhead := sc.getCommitsAhead(remoteHash, localHash)
		if commitsAhead > 0 {
			sc.statusLabel.SetText(fmt.Sprintf("Status: 🟡 Local changes (+%d commits)", commitsAhead))
		} else {
			sc.statusLabel.SetText("Status: 🟡 Local changes")
		}
		sc.actionButton.SetText("Deploy Latest")
	}

	// Set revision labels
	sc.localLabel.SetText(fmt.Sprintf("Local: %s", localHash[:7]))

	// Format deployment time
	timeAgo := formatTimeAgo(lastDeployment.StartedAt)
	sc.remoteLabel.SetText(fmt.Sprintf("Remote: %s (deployed %s)", remoteHash, timeAgo))
}

func (sc *StatusCard) getLocalRevision() string {
	out, err := exec.Command("bash", "-c", "cd "+sc.app.LocalPath+" && git rev-parse HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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

func (sc *StatusCard) SetDeployCallback(callback func()) {
	sc.actionButton.OnTapped = callback
}
