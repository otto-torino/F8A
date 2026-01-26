package models

import (
	"database/sql"
	"errors"
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

func UpdateDeploymentStatus(id int64, status string, errorMessage *string, errorStep *string) error {
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

func UpdateDeploymentDuration(id int64, durationMs *int64) error {
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
	defer stmt.Close()
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
	if err := rows.Err(); err != nil {
		logger.ZapLog.Error("Error iterating deployment rows", err)
		return nil, err
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
	defer stmt.Close()
	err = stmt.QueryRow(appID).Scan(&d.ID, &d.AppID, &d.CommitHash, &d.Status, &d.StartedAt, &completedAt, &d.TotalDurationMs, &d.ErrorMessage, &d.ErrorStep)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if completedAt != nil {
		d.CompletedAt = completedAt
	}
	return &d, nil
}
