package models

import (
	"time"

	"github.com/otto-torino/f8a/db"
	"github.com/otto-torino/f8a/logger"
)

type Deployment struct {
	ID              int        `sql:"id"`
	AppID           int        `sql:"app_id"`
	CommitHash      string     `sql:"commit_hash"`
	Status          string     `sql:"status"`
	StartedAt       time.Time  `sql:"started_at"`
	CompletedAt     *time.Time `sql:"completed_at"`
	TotalDurationMs *int64     `sql:"total_duration_ms"`
	ErrorMessage    *string    `sql:"error_message"`
	ErrorStep       *string    `sql:"error_step"`
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
	id, err := result.LastInsertId()
	if err != nil {
		logger.ZapLog.Error("Cannot get last insert ID", err)
		return 0, err
	}
	return id, nil
}
