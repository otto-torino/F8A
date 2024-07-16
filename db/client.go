package db

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type DBClient struct {
	C *sql.DB
}

var (
	db   *DBClient
	once sync.Once
)

// DB returns a singleton pointer to pgx connection instance
func DB() *DBClient {
	if db == nil {
		once.Do(func() {
			dbFile := filepath.Join(viper.GetString("app.homePath"), "f8a.db")
			conn, err := sql.Open("sqlite3", dbFile)
			if err != nil {
				zap.S().Fatal("Cannot connect to sqlite3 db ", dbFile)
			} else {
				zap.S().Info("Succesfully connected to sqlite3 db on ", dbFile)
			}

			db = &DBClient{conn}
		})
	}

	return db
}

func InitDatabase() {
	dbFile := filepath.Join(viper.GetString("app.homePath"), "f8a.db")
	if _, err := os.Stat(dbFile); errors.Is(err, os.ErrNotExist) {
		os.OpenFile(dbFile, os.O_RDONLY|os.O_CREATE, 0666)
	}

	// sqlite3
	// create a new webapps table if not exists with the following columns:
	// id: auto increment
	// timestamp: timestamp of creation
	// name string
	// local_path: local path
	// remote_host: remote host
	// remote_path: remote path
	// has_htaccess: 0 or 1
	stmt := `
CREATE TABLE IF NOT EXISTS apps (
id INTEGER PRIMARY KEY AUTOINCREMENT,
timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
name TEXT,
local_path TEXT,
remote_host TEXT,
remote_path TEXT,
has_htaccess INTEGER
);
		`
	client := DB()
	_, err := client.C.Exec(stmt)
	if err != nil {
		panic(err)
	}
}
