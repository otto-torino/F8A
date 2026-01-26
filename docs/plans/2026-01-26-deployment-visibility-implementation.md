# Deployment Visibility Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add real-time progress tracking, historical deployment data, and improved status indicators to F8A deployment workflow.

**Architecture:** Three-layer approach: (1) Database layer for deployment history storage, (2) Progress tracking system using channels for real-time updates, (3) Fyne UI components for visualization. Existing deployment flow in commands/deploy.go will be refactored to emit progress events.

**Tech Stack:** Go, Fyne UI, SQLite, existing utils/shell.go for command execution

---

## Task 1: Database Schema - Deployments Table

**Files:**
- Modify: `db/client.go:106` (after InitDatabase function)

**Step 1: Add deployments table creation to InitDatabase**

Add this code after line 105 in `db/client.go`:

```go
	stmt = `
CREATE TABLE IF NOT EXISTS deployments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	app_id INTEGER NOT NULL,
	commit_hash TEXT NOT NULL,
	status TEXT NOT NULL,
	started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	completed_at DATETIME,
	total_duration_ms INTEGER,
	error_message TEXT,
	error_step TEXT,
	FOREIGN KEY (app_id) REFERENCES apps(id)
);
	`
	_, err = client.C.Exec(stmt)
	if err != nil {
		// table might already exist, ignore error
	}

	stmt = `
CREATE INDEX IF NOT EXISTS idx_deployments_app_id_started ON deployments(app_id, started_at DESC);
	`
	_, err = client.C.Exec(stmt)
	if err != nil {
		// index might already exist, ignore error
	}
```

**Step 2: Test database initialization**

Run: `go run main.go` (will create tables on startup)
Expected: Application starts without errors, check logs for "Successfully connected to sqlite3 db"

**Step 3: Commit database schema**

```bash
git add db/client.go
git commit -m "feat(db): add deployments table and index

Add deployments table to track deployment history with app_id, commit_hash,
status, timestamps, duration, and error details. Add index on app_id and
started_at for efficient history queries.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: Database Schema - Deployment Steps Table

**Files:**
- Modify: `db/client.go` (after Task 1 additions)

**Step 1: Add deployment_steps table creation**

Add this code after the deployments table creation in `db/client.go`:

```go
	stmt = `
CREATE TABLE IF NOT EXISTS deployment_steps (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	deployment_id INTEGER NOT NULL,
	step_name TEXT NOT NULL,
	status TEXT NOT NULL,
	started_at DATETIME,
	completed_at DATETIME,
	duration_ms INTEGER,
	output TEXT,
	FOREIGN KEY (deployment_id) REFERENCES deployments(id)
);
	`
	_, err = client.C.Exec(stmt)
	if err != nil {
		// table might already exist, ignore error
	}
```

**Step 2: Test database initialization**

Run: `go run main.go`
Expected: Application starts without errors

**Step 3: Commit deployment steps schema**

```bash
git add db/client.go
git commit -m "feat(db): add deployment_steps table

Add deployment_steps table to track individual step execution within deployments.
Stores step name, status, timing, and output for progress tracking.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: Deployment Models - Create Deployment

**Files:**
- Create: `models/deployment.go`

**Step 1: Create deployment model file**

Create `models/deployment.go`:

```go
package models

import (
	"time"

	"github.com/otto-torino/f8a/db"
	"github.com/otto-torino/f8a/logger"
)

type Deployment struct {
	ID              int       `sql:"id"`
	AppID           int       `sql:"app_id"`
	CommitHash      string    `sql:"commit_hash"`
	Status          string    `sql:"status"`
	StartedAt       time.Time `sql:"started_at"`
	CompletedAt     *time.Time `sql:"completed_at"`
	TotalDurationMs int64     `sql:"total_duration_ms"`
	ErrorMessage    string    `sql:"error_message"`
	ErrorStep       string    `sql:"error_step"`
}

type DeploymentStep struct {
	ID           int        `sql:"id"`
	DeploymentID int        `sql:"deployment_id"`
	StepName     string     `sql:"step_name"`
	Status       string     `sql:"status"`
	StartedAt    *time.Time `sql:"started_at"`
	CompletedAt  *time.Time `sql:"completed_at"`
	DurationMs   int64      `sql:"duration_ms"`
	Output       string     `sql:"output"`
}

func CreateDeployment(appID int, commitHash string) (int64, error) {
	result, err := db.DB().C.Exec(
		"INSERT INTO deployments (app_id, commit_hash, status) VALUES (?, ?, ?)",
		appID, commitHash, "running",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot create deployment", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit deployment creation**

```bash
git add models/deployment.go
git commit -m "feat(models): add deployment creation

Add Deployment and DeploymentStep structs and CreateDeployment function
to track deployment history in database.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: Deployment Models - Update and Query

**Files:**
- Modify: `models/deployment.go`

**Step 1: Add update and query functions**

Add these functions to `models/deployment.go`:

```go
func UpdateDeploymentStatus(id int64, status string, errorMessage string, errorStep string) error {
	now := time.Now()
	_, err := db.DB().C.Exec(
		"UPDATE deployments SET status = ?, completed_at = ?, error_message = ?, error_step = ? WHERE id = ?",
		status, now, errorMessage, errorStep, id,
	)
	if err != nil {
		logger.ZapLog.Error("Cannot update deployment status", err)
		return err
	}
	return nil
}

func UpdateDeploymentDuration(id int64, durationMs int64) error {
	_, err := db.DB().C.Exec(
		"UPDATE deployments SET total_duration_ms = ? WHERE id = ?",
		durationMs, id,
	)
	if err != nil {
		logger.ZapLog.Error("Cannot update deployment duration", err)
		return err
	}
	return nil
}

func GetRecentDeployments(appID int, limit int) ([]Deployment, error) {
	deployments := []Deployment{}
	stmt, err := db.DB().C.Prepare(
		"SELECT id, app_id, commit_hash, status, started_at, completed_at, total_duration_ms, error_message, error_step FROM deployments WHERE app_id = ? ORDER BY started_at DESC LIMIT ?",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot prepare get deployments query", err)
		return nil, err
	}
	rows, err := stmt.Query(appID, limit)
	if err != nil {
		logger.ZapLog.Error("Cannot get deployments", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var d Deployment
		var completedAt *time.Time
		err := rows.Scan(&d.ID, &d.AppID, &d.CommitHash, &d.Status, &d.StartedAt, &completedAt, &d.TotalDurationMs, &d.ErrorMessage, &d.ErrorStep)
		if err != nil {
			logger.ZapLog.Error("Cannot scan deployment row", err)
			continue
		}
		if completedAt != nil {
			d.CompletedAt = completedAt
		}
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func GetLastSuccessfulDeployment(appID int) (*Deployment, error) {
	var d Deployment
	var completedAt *time.Time
	stmt, err := db.DB().C.Prepare(
		"SELECT id, app_id, commit_hash, status, started_at, completed_at, total_duration_ms, error_message, error_step FROM deployments WHERE app_id = ? AND status = 'success' ORDER BY started_at DESC LIMIT 1",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot prepare get last deployment query", err)
		return nil, err
	}
	err = stmt.QueryRow(appID).Scan(&d.ID, &d.AppID, &d.CommitHash, &d.Status, &d.StartedAt, &completedAt, &d.TotalDurationMs, &d.ErrorMessage, &d.ErrorStep)
	if err != nil {
		return nil, err
	}
	if completedAt != nil {
		d.CompletedAt = completedAt
	}
	return &d, nil
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit update and query functions**

```bash
git add models/deployment.go
git commit -m "feat(models): add deployment update and query functions

Add functions to update deployment status and duration, retrieve recent
deployments for an app, and get last successful deployment.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 5: Deployment Step Models

**Files:**
- Modify: `models/deployment.go`

**Step 1: Add deployment step functions**

Add these functions to `models/deployment.go`:

```go
func CreateDeploymentStep(deploymentID int64, stepName string) (int64, error) {
	result, err := db.DB().C.Exec(
		"INSERT INTO deployment_steps (deployment_id, step_name, status) VALUES (?, ?, ?)",
		deploymentID, stepName, "pending",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot create deployment step", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func StartDeploymentStep(stepID int64) error {
	now := time.Now()
	_, err := db.DB().C.Exec(
		"UPDATE deployment_steps SET status = ?, started_at = ? WHERE id = ?",
		"running", now, stepID,
	)
	if err != nil {
		logger.ZapLog.Error("Cannot start deployment step", err)
		return err
	}
	return nil
}

func CompleteDeploymentStep(stepID int64, status string, output string) error {
	now := time.Now()

	// Get start time to calculate duration
	var startedAt time.Time
	err := db.DB().C.QueryRow("SELECT started_at FROM deployment_steps WHERE id = ?", stepID).Scan(&startedAt)
	if err != nil {
		logger.ZapLog.Error("Cannot get step start time", err)
		return err
	}

	durationMs := now.Sub(startedAt).Milliseconds()

	_, err = db.DB().C.Exec(
		"UPDATE deployment_steps SET status = ?, completed_at = ?, duration_ms = ?, output = ? WHERE id = ?",
		status, now, durationMs, output, stepID,
	)
	if err != nil {
		logger.ZapLog.Error("Cannot complete deployment step", err)
		return err
	}
	return nil
}

func GetDeploymentSteps(deploymentID int64) ([]DeploymentStep, error) {
	steps := []DeploymentStep{}
	stmt, err := db.DB().C.Prepare(
		"SELECT id, deployment_id, step_name, status, started_at, completed_at, duration_ms, output FROM deployment_steps WHERE deployment_id = ? ORDER BY id ASC",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot prepare get steps query", err)
		return nil, err
	}
	rows, err := stmt.Query(deploymentID)
	if err != nil {
		logger.ZapLog.Error("Cannot get deployment steps", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s DeploymentStep
		err := rows.Scan(&s.ID, &s.DeploymentID, &s.StepName, &s.Status, &s.StartedAt, &s.CompletedAt, &s.DurationMs, &s.Output)
		if err != nil {
			logger.ZapLog.Error("Cannot scan step row", err)
			continue
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func GetAverageStepDuration(appID int, stepName string, limit int) (int64, error) {
	var avgDuration int64
	err := db.DB().C.QueryRow(
		`SELECT AVG(ds.duration_ms)
		FROM deployment_steps ds
		JOIN deployments d ON ds.deployment_id = d.id
		WHERE d.app_id = ? AND ds.step_name = ? AND ds.status = 'success' AND d.status = 'success'
		ORDER BY d.started_at DESC
		LIMIT ?`,
		appID, stepName, limit,
	).Scan(&avgDuration)
	if err != nil {
		return 0, err
	}
	return avgDuration, nil
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit deployment step functions**

```bash
git add models/deployment.go
git commit -m "feat(models): add deployment step tracking functions

Add functions to create, start, and complete deployment steps. Include
GetDeploymentSteps to retrieve all steps for a deployment and
GetAverageStepDuration for time estimates.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 6: Progress Tracking Package - Types

**Files:**
- Create: `progress/tracker.go`

**Step 1: Create progress package with types**

Create `progress/tracker.go`:

```go
package progress

import (
	"time"
)

type StepStatus int

const (
	StepPending StepStatus = iota
	StepRunning
	StepSuccess
	StepFailed
)

type Step struct {
	ID              int64
	Name            string
	Description     string
	Status          StepStatus
	StartTime       time.Time
	Duration        time.Duration
	AverageDuration time.Duration
	Output          []string
	Error           error
}

type StepUpdate struct {
	StepName string
	Status   StepStatus
	Duration time.Duration
	Output   string
	Error    error
}

type DeploymentProgress struct {
	DeploymentID  int64
	AppID         int
	CommitHash    string
	CurrentStep   string
	Steps         []Step
	StartTime     time.Time
	UpdateChannel chan StepUpdate
}

// Step names constants
const (
	StepBuild    = "build"
	StepArchive  = "archive"
	StepUpload   = "upload"
	StepBackup   = "backup"
	StepExtract  = "extract"
	StepActivate = "activate"
	StepCleanup  = "cleanup"
)

func NewDeploymentProgress(deploymentID int64, appID int, commitHash string) *DeploymentProgress {
	return &DeploymentProgress{
		DeploymentID:  deploymentID,
		AppID:         appID,
		CommitHash:    commitHash,
		StartTime:     time.Now(),
		UpdateChannel: make(chan StepUpdate, 10),
		Steps: []Step{
			{Name: StepBuild, Description: "Building application", Status: StepPending},
			{Name: StepArchive, Description: "Creating archive", Status: StepPending},
			{Name: StepUpload, Description: "Uploading to server", Status: StepPending},
			{Name: StepBackup, Description: "Backing up current version", Status: StepPending},
			{Name: StepExtract, Description: "Extracting archive", Status: StepPending},
			{Name: StepActivate, Description: "Activating new version", Status: StepPending},
			{Name: StepCleanup, Description: "Cleaning up", Status: StepPending},
		},
	}
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit progress types**

```bash
git add progress/tracker.go
git commit -m "feat(progress): add progress tracking types

Add DeploymentProgress, Step, and StepUpdate types for tracking deployment
progress. Define step status constants and step names for all deployment phases.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 7: Progress Tracking Package - Step Execution

**Files:**
- Modify: `progress/tracker.go`

**Step 1: Add step execution methods**

Add these methods to `progress/tracker.go`:

```go
import (
	"github.com/otto-torino/f8a/models"
)

func (dp *DeploymentProgress) StartStep(stepName string) error {
	// Find and update step in memory
	for i := range dp.Steps {
		if dp.Steps[i].Name == stepName {
			dp.Steps[i].Status = StepRunning
			dp.Steps[i].StartTime = time.Now()
			dp.CurrentStep = stepName

			// Create step in database
			stepID, err := models.CreateDeploymentStep(dp.DeploymentID, stepName)
			if err != nil {
				return err
			}
			dp.Steps[i].ID = stepID

			// Start step in database
			err = models.StartDeploymentStep(stepID)
			if err != nil {
				return err
			}

			// Send update
			dp.UpdateChannel <- StepUpdate{
				StepName: stepName,
				Status:   StepRunning,
			}
			break
		}
	}
	return nil
}

func (dp *DeploymentProgress) CompleteStep(stepName string, success bool, output string, err error) error {
	// Find and update step in memory
	for i := range dp.Steps {
		if dp.Steps[i].Name == stepName {
			duration := time.Since(dp.Steps[i].StartTime)
			dp.Steps[i].Duration = duration
			dp.Steps[i].Error = err

			if success {
				dp.Steps[i].Status = StepSuccess
			} else {
				dp.Steps[i].Status = StepFailed
			}

			// Append output
			if output != "" {
				dp.Steps[i].Output = append(dp.Steps[i].Output, output)
			}

			// Update database
			status := "success"
			if !success {
				status = "failed"
			}
			dbErr := models.CompleteDeploymentStep(dp.Steps[i].ID, status, output)
			if dbErr != nil {
				return dbErr
			}

			// Send update
			dp.UpdateChannel <- StepUpdate{
				StepName: stepName,
				Status:   dp.Steps[i].Status,
				Duration: duration,
				Output:   output,
				Error:    err,
			}
			break
		}
	}
	return nil
}

func (dp *DeploymentProgress) GetProgress() float64 {
	completed := 0
	for _, step := range dp.Steps {
		if step.Status == StepSuccess {
			completed++
		}
	}
	return float64(completed) / float64(len(dp.Steps)) * 100
}

func (dp *DeploymentProgress) LoadAverageDurations() error {
	for i := range dp.Steps {
		avgDuration, err := models.GetAverageStepDuration(dp.AppID, dp.Steps[i].Name, 10)
		if err == nil && avgDuration > 0 {
			dp.Steps[i].AverageDuration = time.Duration(avgDuration) * time.Millisecond
		}
	}
	return nil
}

func (dp *DeploymentProgress) Close() {
	close(dp.UpdateChannel)
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit step execution methods**

```bash
git add progress/tracker.go
git commit -m "feat(progress): add step execution tracking

Add StartStep and CompleteStep methods to track step lifecycle. Include
GetProgress for percentage calculation and LoadAverageDurations for
time estimates from historical data.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 8: Refactor Deploy Command - Setup Progress Tracker

**Files:**
- Modify: `commands/deploy.go:19-36` (Deploy function)

**Step 1: Update imports and Deploy function**

Replace the Deploy function in `commands/deploy.go`:

```go
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

func Deploy(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		outputContainer.RemoveAll()

		// Get commit hash
		out, err := exec.Command("bash", "-c", "cd "+app.LocalPath+" && git rev-parse --short HEAD").Output()
		if err != nil {
			utils.AddTextToOutput("Failed to get commit hash: "+err.Error(), errorColor, outputContainer)
			return
		}
		commitHash := strings.TrimSpace(string(out))

		utils.AddTextToOutput("Deploying revision "+commitHash, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)

		// Create deployment record
		deploymentID, err := models.CreateDeployment(app.ID, commitHash)
		if err != nil {
			utils.AddTextToOutput("Failed to create deployment record: "+err.Error(), errorColor, outputContainer)
			return
		}

		// Initialize progress tracker
		tracker := progress.NewDeploymentProgress(deploymentID, app.ID, commitHash)
		tracker.LoadAverageDurations()
		defer tracker.Close()

		// Start deployment
		startTime := time.Now()
		err = deployWithProgress(app, outputContainer, commitHash, tracker)
		duration := time.Since(startTime)

		if err != nil {
			utils.AddTextToOutput("Deployment failed for revision "+commitHash, errorColor, outputContainer)
			models.UpdateDeploymentStatus(deploymentID, "failed", err.Error(), tracker.CurrentStep)
			return
		}

		utils.AddTextToOutput("Deployed revision "+commitHash, color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
		models.UpdateDeploymentStatus(deploymentID, "success", "", "")
		models.UpdateDeploymentDuration(deploymentID, duration.Milliseconds())
	}
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit deploy function refactor**

```bash
git add commands/deploy.go
git commit -m "refactor(deploy): integrate progress tracking

Update Deploy function to create deployment records and initialize progress
tracker. Track deployment duration and status in database.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 9: Refactor Deploy Command - Wrap Steps with Progress

**Files:**
- Modify: `commands/deploy.go:38-87` (deploy function, rename to deployWithProgress)

**Step 1: Rename and refactor deploy function**

Replace the `deploy` function in `commands/deploy.go` with `deployWithProgress`:

```go
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
		err = utils.Shellout(fmt.Sprintf("ssh otto@%s mv %s/%s %s/%s", app.RemoteHost, app.RemotePath, app.LocalDistDirName, app.RemotePath, commitHash), outputContainer, false)
	}
	tracker.CompleteStep(progress.StepExtract, err == nil, "", err)
	if err != nil {
		return err
	}

	// Skip ls command (not part of critical path)
	utils.Shellout(fmt.Sprintf("ssh otto@%s ls -la %s", app.RemoteHost, app.RemotePath), outputContainer, false)

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
```

**Step 2: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit progress-wrapped deployment**

```bash
git add commands/deploy.go
git commit -m "refactor(deploy): wrap steps with progress tracking

Rename deploy to deployWithProgress and wrap each deployment step with
StartStep/CompleteStep calls to track progress in database.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 10: Status Card Component - Basic Structure

**Files:**
- Create: `components/status_card.go`

**Step 1: Create status card component**

Create `components/status_card.go`:

```go
package components

import (
	"fmt"
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
	background := canvas.NewRectangle(canvas.NewLinearGradientColor(
		&fyne.LinearGradient{
			StartColor: fyne.NewColor(240, 240, 240),
			EndColor:   fyne.NewColor(250, 250, 250),
			Angle:      90,
		},
	))

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
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit status card structure**

```bash
git add components/status_card.go
git commit -m "feat(components): add status card basic structure

Create StatusCard widget with labels for status, remote/local revisions,
and action button. Implement custom renderer with gradient background.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 11: Status Card Component - Update Logic

**Files:**
- Modify: `components/status_card.go`

**Step 1: Add UpdateStatus method**

Add this method to `components/status_card.go`:

```go
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
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit status update logic**

```bash
git add components/status_card.go
git commit -m "feat(components): add status card update logic

Add UpdateStatus method to compare local and remote revisions, calculate
commits ahead, and format time ago. Include SetDeployCallback for
action button integration.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 12: Integrate Status Card into App Content

**Files:**
- Modify: `components/app_content.go:123-174` (HandleWebAppSection function)

**Step 1: Add status card to app detail view**

Replace the HandleWebAppSection function in `components/app_content.go`:

```go
func HandleWebAppSection(id int) {
	app, err := models.GetApp(id)
	if err != nil {
		zap.S().Error("Cannot get app", err)
		return
	}

	// clean
	mainContent.RemoveAll()

	// Status card at top
	statusCard := NewStatusCard(app)
	statusCard.UpdateStatus()

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
	localDistDirNameLabel := widget.NewLabel("App Local Dist Dir Name")
	localDistDirName := widget.NewLabel(app.LocalDistDirName)
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
	infoGrid := container.New(layout.NewFormLayout(), nameLabel, name, localPathLabel, localPath, localDistDirNameLabel, localDistDirName, remoteHostLabel, remoteHost, remotePathLabel, remotePath, currentDirNameLabel, currentDirName, hasHtAccessLabel, hasHtAccess)

	top := container.NewVBox(statusCard, header, infoGrid)

	output := container.NewVBox()
	background := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	utils.Scroll = container.NewScroll(container.New(layout.NewPaddedLayout(), output))
	outputContainer := container.NewBorder(nil, nil, nil, nil, container.NewStack(background, utils.Scroll))

	actionButtons := MakeActionButtons(app, output)

	mainContent.Add(container.NewBorder(top, actionButtons, nil, nil, outputContainer))
}
```

**Step 2: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit status card integration**

```bash
git add components/app_content.go
git commit -m "feat(components): integrate status card into app detail view

Add StatusCard to top of app detail view to show deployment status and
revision comparison at a glance.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 13: History Timeline Component - Basic Structure

**Files:**
- Create: `components/history_timeline.go`

**Step 1: Create history timeline component**

Create `components/history_timeline.go`:

```go
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
	statusIcon := "✓"
	if d.Status == "failed" {
		statusIcon = "✗"
	} else if d.Status == "running" {
		statusIcon = "⟳"
	}

	// Commit hash
	commitLabel := widget.NewLabel(fmt.Sprintf("%s %s", statusIcon, d.CommitHash))

	// Time ago
	timeAgo := formatTimeAgo(d.StartedAt)
	timeLabel := widget.NewLabel(timeAgo)

	// Duration or error
	var durationLabel *widget.Label
	if d.Status == "success" && d.TotalDurationMs > 0 {
		duration := time.Duration(d.TotalDurationMs) * time.Millisecond
		durationLabel = widget.NewLabel(fmt.Sprintf("(%s)", formatDuration(duration)))
	} else if d.Status == "failed" {
		durationLabel = widget.NewLabel(fmt.Sprintf("failed at %s", d.ErrorStep))
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
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit history timeline structure**

```bash
git add components/history_timeline.go
git commit -m "feat(components): add history timeline basic structure

Create HistoryTimeline widget to display recent deployments with status,
commit hash, timestamp, and duration. Include collapsible header.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 14: History Timeline - Detail Dialog

**Files:**
- Modify: `components/history_timeline.go`

**Step 1: Implement detail dialog**

Replace the `showDeploymentDetails` placeholder in `components/history_timeline.go`:

```go
func (ht *HistoryTimeline) showDeploymentDetails(d models.Deployment) {
	// Get deployment steps
	steps, err := models.GetDeploymentSteps(int64(d.ID))
	if err != nil {
		return
	}

	// Build content
	content := container.NewVBox()

	// Header info
	content.Add(widget.NewLabel(fmt.Sprintf("Deployment: %s", d.CommitHash)))
	content.Add(widget.NewLabel(fmt.Sprintf("Started: %s", d.StartedAt.Format("2006-01-02 15:04:05"))))

	if d.CompletedAt != nil {
		content.Add(widget.NewLabel(fmt.Sprintf("Completed: %s", d.CompletedAt.Format("2006-01-02 15:04:05"))))
	}

	content.Add(widget.NewLabel(fmt.Sprintf("Status: %s", d.Status)))

	if d.TotalDurationMs > 0 {
		duration := time.Duration(d.TotalDurationMs) * time.Millisecond
		content.Add(widget.NewLabel(fmt.Sprintf("Duration: %s", formatDuration(duration))))
	}

	if d.ErrorMessage != "" {
		content.Add(widget.NewLabel(fmt.Sprintf("Error: %s", d.ErrorMessage)))
	}

	// Spacer
	content.Add(widget.NewLabel(""))
	content.Add(widget.NewLabel("Steps:"))

	// Steps list
	for _, step := range steps {
		statusIcon := "✓"
		if step.Status == "failed" {
			statusIcon = "✗"
		} else if step.Status == "running" {
			statusIcon = "⟳"
		} else if step.Status == "pending" {
			statusIcon = "--"
		}

		stepDuration := ""
		if step.DurationMs > 0 {
			stepDuration = formatDuration(time.Duration(step.DurationMs) * time.Millisecond)
		}

		stepLabel := fmt.Sprintf("%s %s    %s", statusIcon, step.StepName, stepDuration)
		content.Add(widget.NewLabel(stepLabel))
	}

	// Show dialog
	registry := utils.Registry()
	scrollContent := container.NewScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("Deployment Details", "Close", scrollContent, *registry.Window)
	d.Resize(fyne.NewSize(550, 500))
	d.Show()
}
```

Add import at top:

```go
import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/models"
	"github.com/otto-torino/f8a/utils"
)
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit detail dialog**

```bash
git add components/history_timeline.go
git commit -m "feat(components): add deployment detail dialog

Implement showDeploymentDetails to display full deployment information
including all steps with timing and status in a dialog.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 15: Integrate History Timeline into App Content

**Files:**
- Modify: `components/app_content.go` (HandleWebAppSection function)

**Step 1: Add history timeline below action buttons**

Update the bottom of `HandleWebAppSection` function in `components/app_content.go`:

```go
	actionButtons := MakeActionButtons(app, output)

	// Add history timeline
	historyTimeline := NewHistoryTimeline(app)

	mainContent.Add(container.NewBorder(top, container.NewVBox(actionButtons, historyTimeline), nil, nil, outputContainer))
}
```

**Step 2: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit history timeline integration**

```bash
git add components/app_content.go
git commit -m "feat(components): integrate history timeline into app view

Add HistoryTimeline below action buttons to show recent deployment history
with expandable details.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 16: Desktop Notifications - Add Notification Utility

**Files:**
- Create: `utils/notifications.go`

**Step 1: Create notifications utility**

Create `utils/notifications.go`:

```go
package utils

import (
	"time"

	"fyne.io/fyne/v2"
)

func SendSuccessNotification(appName string, commitHash string, duration time.Duration) {
	registry := Registry()
	if registry.Application == nil {
		return
	}

	notification := fyne.NewNotification(
		"F8A Deployment Complete",
		appName+" ("+commitHash+") deployed successfully in "+formatDurationShort(duration),
	)
	(*registry.Application).SendNotification(notification)
}

func SendFailureNotification(appName string, errorStep string) {
	registry := Registry()
	if registry.Application == nil {
		return
	}

	notification := fyne.NewNotification(
		"F8A Deployment Failed",
		appName+" deployment failed at "+errorStep+" step",
	)
	(*registry.Application).SendNotification(notification)
}

func formatDurationShort(d time.Duration) string {
	if d < time.Second {
		return "less than a second"
	} else if d < time.Minute {
		return string(int(d.Seconds())) + "s"
	} else {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs == 0 {
			return string(mins) + "m"
		}
		return string(mins) + "m " + string(secs) + "s"
	}
}
```

**Step 2: Build to verify syntax**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit notification utility**

```bash
git add utils/notifications.go
git commit -m "feat(utils): add desktop notification support

Add SendSuccessNotification and SendFailureNotification functions to
send desktop notifications on deployment completion or failure.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 17: Integrate Notifications into Deploy Command

**Files:**
- Modify: `commands/deploy.go` (Deploy function)

**Step 1: Add notifications to deployment flow**

Update the Deploy function in `commands/deploy.go` to send notifications:

```go
func Deploy(app *models.App, outputContainer *fyne.Container) func() {
	return func() {
		outputContainer.RemoveAll()

		// Get commit hash
		out, err := exec.Command("bash", "-c", "cd "+app.LocalPath+" && git rev-parse --short HEAD").Output()
		if err != nil {
			utils.AddTextToOutput("Failed to get commit hash: "+err.Error(), errorColor, outputContainer)
			return
		}
		commitHash := strings.TrimSpace(string(out))

		utils.AddTextToOutput("Deploying revision "+commitHash, color.RGBA{R: 255, G: 153, B: 0, A: 255}, outputContainer)

		// Create deployment record
		deploymentID, err := models.CreateDeployment(app.ID, commitHash)
		if err != nil {
			utils.AddTextToOutput("Failed to create deployment record: "+err.Error(), errorColor, outputContainer)
			return
		}

		// Initialize progress tracker
		tracker := progress.NewDeploymentProgress(deploymentID, app.ID, commitHash)
		tracker.LoadAverageDurations()
		defer tracker.Close()

		// Start deployment
		startTime := time.Now()
		err = deployWithProgress(app, outputContainer, commitHash, tracker)
		duration := time.Since(startTime)

		if err != nil {
			utils.AddTextToOutput("Deployment failed for revision "+commitHash, errorColor, outputContainer)
			models.UpdateDeploymentStatus(deploymentID, "failed", err.Error(), tracker.CurrentStep)
			utils.SendFailureNotification(app.Name, tracker.CurrentStep)
			return
		}

		utils.AddTextToOutput("Deployed revision "+commitHash, color.RGBA{R: 0, G: 255, B: 0, A: 255}, outputContainer)
		models.UpdateDeploymentStatus(deploymentID, "success", "", "")
		models.UpdateDeploymentDuration(deploymentID, duration.Milliseconds())
		utils.SendSuccessNotification(app.Name, commitHash, duration)
	}
}
```

**Step 2: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit notification integration**

```bash
git add commands/deploy.go
git commit -m "feat(deploy): integrate desktop notifications

Send success/failure notifications on deployment completion. Includes
deployment duration and error details.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 18: Update Status Card After Deployment

**Files:**
- Modify: `components/app_content.go` (MakeActionButtons function)

**Step 1: Store status card reference and update after deploy**

Modify `HandleWebAppSection` in `components/app_content.go`:

```go
func HandleWebAppSection(id int) {
	app, err := models.GetApp(id)
	if err != nil {
		zap.S().Error("Cannot get app", err)
		return
	}

	// clean
	mainContent.RemoveAll()

	// Status card at top
	statusCard := NewStatusCard(app)
	statusCard.UpdateStatus()

	// History timeline
	historyTimeline := NewHistoryTimeline(app)

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
	localDistDirNameLabel := widget.NewLabel("App Local Dist Dir Name")
	localDistDirName := widget.NewLabel(app.LocalDistDirName)
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
	infoGrid := container.New(layout.NewFormLayout(), nameLabel, name, localPathLabel, localPath, localDistDirNameLabel, localDistDirName, remoteHostLabel, remoteHost, remotePathLabel, remotePath, currentDirNameLabel, currentDirName, hasHtAccessLabel, hasHtAccess)

	top := container.NewVBox(statusCard, header, infoGrid)

	output := container.NewVBox()
	background := canvas.NewRectangle(color.RGBA{R: 0, G: 0, B: 0, A: 255})
	utils.Scroll = container.NewScroll(container.New(layout.NewPaddedLayout(), output))
	outputContainer := container.NewBorder(nil, nil, nil, nil, container.NewStack(background, utils.Scroll))

	actionButtons := MakeActionButtons(app, output, statusCard, historyTimeline)

	mainContent.Add(container.NewBorder(top, container.NewVBox(actionButtons, historyTimeline), nil, nil, outputContainer))
}
```

**Step 2: Update MakeActionButtons signature and deploy callback**

Modify `MakeActionButtons` in `components/app_content.go`:

```go
func MakeActionButtons(app *models.App, outputContainer *fyne.Container, statusCard *StatusCard, historyTimeline *HistoryTimeline) *fyne.Container {
	// Build Button
	build := utils.MakeButton("Build", commands.Build(app, outputContainer))

	// Build Archive
	buildArchive := utils.MakeButton("Build Archive", commands.BuildArchive(app, outputContainer))

	// Local revision button
	gitLocaleRev := utils.MakeButton("Local Revision", commands.LocalRevision(app, outputContainer))

	// Remote revision button
	gitRemoteRev := utils.MakeButton("Remote Revision", commands.RemoteRevision(app, outputContainer))

	// Deploy button with callback to refresh status
	deployCallback := commands.Deploy(app, outputContainer)
	deployWithRefresh := func() {
		deployCallback()
		// Refresh status card and history after deployment
		statusCard.UpdateStatus()
		historyTimeline.LoadDeployments()
	}
	deploy := utils.MakeButton("Deploy", deployWithRefresh)

	// Restore button
	restore := utils.MakeButton("Restore Revision", commands.Restore(app))

	return container.NewHBox(build, buildArchive, gitLocaleRev, gitRemoteRev, restore, deploy)
}
```

**Step 3: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 4: Commit status refresh**

```bash
git add components/app_content.go
git commit -m "feat(components): refresh status and history after deployment

Update status card and history timeline after deployment completes to
reflect new deployment state.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 19: Fix Notification Duration Formatting

**Files:**
- Modify: `utils/notifications.go`

**Step 1: Fix formatDurationShort function**

Replace `formatDurationShort` in `utils/notifications.go`:

```go
func formatDurationShort(d time.Duration) string {
	if d < time.Second {
		return "less than a second"
	} else if d < time.Minute {
		secs := int(d.Seconds())
		return fmt.Sprintf("%ds", secs)
	} else {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		if secs == 0 {
			return fmt.Sprintf("%dm", mins)
		}
		return fmt.Sprintf("%dm %ds", mins, secs)
	}
}
```

Add import at top:

```go
import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
)
```

**Step 2: Build and test**

Run: `go build`
Expected: Build succeeds without errors

**Step 3: Commit duration fix**

```bash
git add utils/notifications.go
git commit -m "fix(utils): correct duration formatting in notifications

Fix formatDurationShort to properly format minutes and seconds using
fmt.Sprintf instead of string conversion.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 20: Documentation - Update README

**Files:**
- Modify: `README.md`

**Step 1: Add deployment visibility features to README**

Add this section after the Install section in `README.md`:

```markdown
## Features

### Deployment Tracking

F8A tracks all deployments with detailed progress information:

- **Real-time Progress**: See which step is currently running during deployment
- **Time Estimates**: View average duration for each step based on historical data
- **Deployment History**: Access last 10 deployments with status, duration, and commit details
- **Status Dashboard**: Compare local vs deployed revisions at a glance
- **Desktop Notifications**: Get notified when deployments complete or fail

### Deployment Steps

Each deployment consists of 7 tracked steps:
1. Build - Run yarn build locally
2. Archive - Create tar file
3. Upload - SCP to remote server
4. Backup - Move current version to previous
5. Extract - Extract new build
6. Activate - Create symlink
7. Cleanup - Remove tar and copy .htaccess if needed
```

**Step 2: Commit README update**

```bash
git add README.md
git commit -m "docs: document deployment visibility features

Add documentation for deployment tracking, progress monitoring, and
history features in README.

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Summary

This implementation plan adds comprehensive deployment visibility to F8A:

1. **Database layer** (Tasks 1-5): Deployments and deployment_steps tables with models
2. **Progress tracking** (Tasks 6-9): Real-time step tracking with channel-based updates
3. **Status card** (Tasks 10-12): Shows deployment status and revision comparison
4. **History timeline** (Tasks 13-15): Recent deployments with expandable details
5. **Notifications** (Tasks 16-17): Desktop alerts on completion/failure
6. **Integration** (Tasks 18-20): Connect all components and update documentation

Each task is bite-sized (2-5 minutes) with clear steps: implementation, build verification, and commit.

---

## Testing Checklist

After implementation, test these scenarios:

- [ ] Deploy an app and verify progress is tracked in database
- [ ] Check status card shows correct local vs remote revision comparison
- [ ] Verify history timeline displays recent deployments
- [ ] Click "View details" on a deployment to see step breakdown
- [ ] Trigger a deployment failure (disconnect network) and verify error is captured
- [ ] Confirm desktop notification appears on deployment completion
- [ ] Verify time estimates appear after multiple deployments
- [ ] Test with app that has never been deployed (should show "Never deployed")
- [ ] Verify status card updates after deployment completes
- [ ] Check history timeline refreshes after new deployment
