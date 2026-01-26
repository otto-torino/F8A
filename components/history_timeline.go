package components

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
)

type HistoryTimeline struct {
	widget.BaseWidget
	app       *models.App
	container *fyne.Container
	expanded  bool
}

func NewHistoryTimeline(app *models.App) *HistoryTimeline {
	timeline := &HistoryTimeline{
		app:       app,
		container: container.NewVBox(),
		expanded:  true,
	}
	timeline.ExtendBaseWidget(timeline)
	timeline.LoadDeployments()
	return timeline
}

func (ht *HistoryTimeline) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ht.container)
}

func (ht *HistoryTimeline) LoadDeployments() {
	ht.container.Objects = nil

	// Header with toggle
	header := widget.NewButton("Recent Deployments ▼", func() {
		ht.Toggle()
	})
	ht.container.Add(header)

	if !ht.expanded {
		ht.container.Refresh()
		return
	}

	// Get recent deployments
	deployments, err := models.GetRecentDeployments(ht.app.ID, 10)
	if err != nil || len(deployments) == 0 {
		ht.container.Add(widget.NewLabel("No deployment history yet"))
		ht.container.Refresh()
		return
	}

	// Display each deployment
	for _, d := range deployments {
		ht.container.Add(ht.makeDeploymentEntry(d))
	}

	ht.container.Refresh()
}

func (ht *HistoryTimeline) makeDeploymentEntry(d models.Deployment) *fyne.Container {
	// Status icon
	statusIcon := "[OK]"
	if d.Status == "failed" {
		statusIcon = "[FAIL]"
	} else if d.Status == "running" {
		statusIcon = "[...]"
	}

	// Commit hash
	commitLabel := widget.NewLabel(fmt.Sprintf("%s %s", statusIcon, d.CommitHash))

	// Time ago
	timeAgo := formatTimeAgo(d.StartedAt)
	timeLabel := widget.NewLabel(timeAgo)

	// Duration or error
	var durationLabel *widget.Label
	if d.Status == "success" && d.TotalDurationMs != nil && *d.TotalDurationMs > 0 {
		duration := time.Duration(*d.TotalDurationMs) * time.Millisecond
		durationLabel = widget.NewLabel(fmt.Sprintf("(%s)", formatDuration(duration)))
	} else if d.Status == "failed" && d.ErrorStep != nil {
		durationLabel = widget.NewLabel(fmt.Sprintf("failed at %s", *d.ErrorStep))
	} else {
		durationLabel = widget.NewLabel("running...")
	}

	// View details button
	detailsBtn := widget.NewButton("View details", func() {
		ht.showDeploymentDetails(d)
	})

	entry := container.NewHBox(
		commitLabel,
		timeLabel,
		durationLabel,
		detailsBtn,
	)

	return entry
}

func (ht *HistoryTimeline) Toggle() {
	ht.expanded = !ht.expanded
	ht.LoadDeployments()
}

func (ht *HistoryTimeline) showDeploymentDetails(d models.Deployment) {
	// Placeholder for detail dialog (will implement in next task)
	fmt.Println("Show details for deployment", d.ID)
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
}
