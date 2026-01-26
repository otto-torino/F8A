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
