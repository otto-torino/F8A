package components

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/progress"
)

type StepRow struct {
	widget.BaseWidget
	step            progress.Step
	icon            *widget.Icon
	nameLabel       *widget.Label
	timeLabel       *widget.Label
	avgLabel        *widget.Label
	toggleBtn       *widget.Button
	outputArea      *widget.Entry
	outputContainer *fyne.Container
	outputVisible   bool
}

func newStepRow(step progress.Step) *StepRow {
	sr := &StepRow{
		step:      step,
		icon:      widget.NewIcon(theme.RadioButtonIcon()),
		nameLabel: widget.NewLabel(step.Description),
		timeLabel: widget.NewLabel("--"),
		avgLabel:  widget.NewLabel(fmt.Sprintf("(avg: %s)", formatStepDuration(step.AverageDuration))),
	}

	sr.outputArea = widget.NewMultiLineEntry()
	sr.outputArea.Disable()
	sr.outputContainer = container.NewVBox(sr.outputArea)
	sr.outputContainer.Hide()

	sr.toggleBtn = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		sr.outputVisible = !sr.outputVisible
		if sr.outputVisible {
			sr.toggleBtn.SetIcon(theme.MenuDropDownIcon())
			sr.outputContainer.Show()
		} else {
			sr.toggleBtn.SetIcon(theme.NavigateNextIcon())
			sr.outputContainer.Hide()
		}
	})
	sr.toggleBtn.Disable()

	sr.ExtendBaseWidget(sr)
	return sr
}

func (sr *StepRow) CreateRenderer() fyne.WidgetRenderer {
	topContainer := container.NewHBox(
		sr.toggleBtn,
		sr.icon,
		sr.nameLabel,
		layout.NewSpacer(),
		sr.timeLabel,
		sr.avgLabel,
	)
	content := container.NewVBox(
		topContainer,
		sr.outputContainer,
	)
	return widget.NewSimpleRenderer(content)
}

func (sr *StepRow) update(update progress.StepUpdate) {
	sr.step.Status = update.Status
	if update.Duration > 0 {
		sr.step.Duration = update.Duration
	}

	if update.Output != "" {
		sr.toggleBtn.Enable()
		sr.outputArea.SetText(sr.outputArea.Text + update.Output)
	}

	switch update.Status {
	case progress.StepRunning:
		sr.icon.SetResource(theme.ViewRefreshIcon())
	case progress.StepSuccess:
		sr.icon.SetResource(theme.ConfirmIcon())
		sr.timeLabel.SetText(formatStepDuration(sr.step.Duration))
	case progress.StepFailed:
		sr.icon.SetResource(theme.ErrorIcon())
		sr.timeLabel.SetText(formatStepDuration(sr.step.Duration))
	}
	sr.Refresh()
}

type ProgressPanel struct {
	widget.BaseWidget
	appID          int
	container      *fyne.Container
	progressBar    *widget.ProgressBar
	stepRows       map[string]*StepRow
	rowsContainer  *fyne.Container
	totalTimeLabel *widget.Label
	startTime      time.Time
}

func NewProgressPanel(appID int, steps []progress.Step) *ProgressPanel {
	pp := &ProgressPanel{
		appID:         appID,
		stepRows:      make(map[string]*StepRow),
		rowsContainer: container.NewVBox(),
		progressBar:   widget.NewProgressBar(),
	}
	pp.ExtendBaseWidget(pp)

	for _, step := range steps {
		row := newStepRow(step)
		pp.stepRows[step.Name] = row
		pp.rowsContainer.Add(row)
	}

	pp.totalTimeLabel = widget.NewLabel("Total time: 0s")
	pp.container = container.NewVBox(
		widget.NewLabelWithStyle("Deployment Progress", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		pp.progressBar,
		pp.rowsContainer,
		pp.totalTimeLabel,
	)

	return pp
}

func (pp *ProgressPanel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pp.container)
}

func (pp *ProgressPanel) Listen(updateChan chan progress.StepUpdate) {
	pp.startTime = time.Now()
	go func() {
		for update := range updateChan {
			row, ok := pp.stepRows[update.StepName]
			if ok {
				row.update(update)
				pp.updateOverallProgress()
			}
		}
	}()

	// Timer goroutine for total time
	go func() {
		for {
			time.Sleep(time.Second)
			// Ideally we'd have a 'Done' signal, but for now we just update
			pp.totalTimeLabel.SetText(fmt.Sprintf("Total time: %s", time.Since(pp.startTime).Round(time.Second)))
		}
	}()
}

func (pp *ProgressPanel) updateOverallProgress() {
	completed := 0
	total := len(pp.stepRows)
	if total == 0 {
		return
	}
	for _, row := range pp.stepRows {
		if row.step.Status == progress.StepSuccess {
			completed++
		}
	}
	pp.progressBar.SetValue(float64(completed) / float64(total))
}

func formatStepDuration(d time.Duration) string {
	if d == 0 {
		return "--"
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
