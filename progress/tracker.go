package progress

import (
	"time"

	"github.com/otto-torino/f8a/models"
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
