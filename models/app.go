package models

import (
	"github.com/otto-torino/f8a/db"
	"github.com/otto-torino/f8a/logger"
)

type App struct {
	ID               int    `sql:"id"`
	Name             string `sql:"name"`
	LocalPath        string `sql:"local_path"`
	LocalDistDirName string `sql:"local_dist_dir_name"`
	RemoteHost       string `sql:"remote_host"`
	RemotePath       string `sql:"remote_path"`
	CurrentDirName   string `sql:"current_dir_name"`
	HasHtAccess      int    `sql:"has_htaccess"`
	HealthCheckUrl   string `sql:"health_check_url"`
}

func CreateApp(name string, localPath string, localDistDirName string, remoteHost string, remotePath string, currentDirName string, hasHtAccess bool, healthCheckUrl string) (int64, error) {
	hasHtAccessInt := 0
	if hasHtAccess {
		hasHtAccessInt = 1
	}
	result, err := db.DB().C.Exec("INSERT INTO apps (name, local_path, local_dist_dir_name, remote_host, remote_path, current_dir_name, has_htaccess, health_check_url) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", name, localPath, localDistDirName, remoteHost, remotePath, currentDirName, hasHtAccessInt, healthCheckUrl)
	if err != nil {
		logger.ZapLog.Error("Cannot create app", err)
		return 0, err
	}
	lastId, _ := result.LastInsertId()
	return lastId, nil
}

func UpdateApp(id int, name string, localPath string, localDistDirName string, remoteHost string, remotePath string, currentDirName string, hasHtAccess bool, healthCheckUrl string) error {
	hasHtAccessInt := 0
	if hasHtAccess {
		hasHtAccessInt = 1
	}
	_, err := db.DB().C.Exec("UPDATE apps SET name = ?, local_path = ?, local_dist_dir_name = ?, remote_host = ?, remote_path = ?, current_dir_name = ?, has_htaccess = ?, health_check_url = ? WHERE id = ?", name, localPath, localDistDirName, remoteHost, remotePath, currentDirName, hasHtAccessInt, healthCheckUrl, id)
	if err != nil {
		logger.ZapLog.Error("Cannot update app", err)
		return err
	}
	return nil
}

func GetApps() ([]App, error) {
	apps := []App{}
	stm, err := db.DB().C.Prepare("SELECT id, name, local_path, local_dist_dir_name, remote_host, remote_path, current_dir_name, has_htaccess, health_check_url FROM apps")
	if err != nil {
		logger.ZapLog.Error("Cannot get apps", err)
		return nil, err
	}
	rows, err := stm.Query()
	if err != nil {
		logger.ZapLog.Error("Cannot get apps", err)
		return nil, err
	} else {
		for rows.Next() {
			var app App
			rows.Scan(&app.ID, &app.Name, &app.LocalPath, &app.LocalDistDirName, &app.RemoteHost, &app.RemotePath, &app.CurrentDirName, &app.HasHtAccess, &app.HealthCheckUrl)
			apps = append(apps, app)
		}
		return apps, nil
	}
}

func GetApp(id int) (*App, error) {
	app := App{}
	stm, err := db.DB().C.Prepare("SELECT id, name, local_path, local_dist_dir_name, remote_host, remote_path, current_dir_name, has_htaccess, health_check_url FROM apps WHERE id = ?")
	if err != nil {
		logger.ZapLog.Error("Cannot get app", err)
		return nil, err
	}
	err = stm.QueryRow(id).Scan(&app.ID, &app.Name, &app.LocalPath, &app.LocalDistDirName, &app.RemoteHost, &app.RemotePath, &app.CurrentDirName, &app.HasHtAccess, &app.HealthCheckUrl)
	if err != nil {
		logger.ZapLog.Error("Cannot get app", err)
		return nil, err
	}
	return &app, nil
}

func (a *App) Delete() error {
	_, err := db.DB().C.Exec("DELETE FROM apps WHERE id = ?", a.ID)
	return err
}
