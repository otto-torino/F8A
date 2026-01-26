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

func CreateDeploymentStep(deploymentID int64, stepName string) (int64, error) {
	result, err := db.DB().C.Exec(
		"INSERT INTO deployment_steps (deployment_id, step_name, status) VALUES (?, ?, ?)",
		deploymentID, stepName, "pending",
	)
	if err != nil {
		logger.ZapLog.Error("Cannot create deployment step", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		logger.ZapLog.Error("Cannot get last insert ID", err)
		return 0, err
	}
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
	defer stmt.Close()
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
	if err = rows.Err(); err != nil {
		logger.ZapLog.Error("Error during rows iteration", err)
		return nil, err
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
