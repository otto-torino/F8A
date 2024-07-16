package models

import (
	"github.com/otto-torino/f8a/db"
	"github.com/otto-torino/f8a/logger"
)

type App struct {
	ID          int    `sql:"id"`
	Name        string `sql:"name"`
	LocalPath   string `sql:"local_path"`
	RemoteHost  string `sql:"remote_host"`
	RemotePath  string `sql:"remote_path"`
	HasHtAccess int    `sql:"has_htaccess"`
}

func CreateApp(name string, localPath string, remoteHost string, remotePath string, hasHtAccess bool) (int64, error) {
	hasHtAccessInt := 0
	if hasHtAccess {
		hasHtAccessInt = 1
	}
	result, err := db.DB().C.Exec("INSERT INTO apps (name, local_path, remote_host, remote_path, has_htaccess) VALUES (?, ?, ?, ?, ?)", name, localPath, remoteHost, remotePath, hasHtAccessInt)
	if err != nil {
		logger.ZapLog.Error("Cannot create app", err)
		return 0, err
	}
	lastId, _ := result.LastInsertId()
	return lastId, nil
}

func GetApps() ([]App, error) {
	apps := []App{}
	stm, err := db.DB().C.Prepare("SELECT id, name, local_path, remote_host, remote_path, has_htaccess FROM apps")
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
			rows.Scan(&app.ID, &app.Name, &app.LocalPath, &app.RemoteHost, &app.RemotePath, &app.HasHtAccess)
			apps = append(apps, app)
		}
		return apps, nil
	}
}

func GetApp(id int) (*App, error) {
	app := App{}
	stm, err := db.DB().C.Prepare("SELECT id, name, local_path, remote_host, remote_path, has_htaccess FROM apps WHERE id = ?")
	if err != nil {
		logger.ZapLog.Error("Cannot get app", err)
		return nil, err
	}
	err = stm.QueryRow(id).Scan(&app.ID, &app.Name, &app.LocalPath, &app.RemoteHost, &app.RemotePath, &app.HasHtAccess)
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
