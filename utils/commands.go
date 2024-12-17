package utils

import (
	"fmt"
	"strings"

	"github.com/otto-torino/f8a/models"
)

func CheckRemoteRevisionEqualsLocal(app *models.App) bool {
	out, err := Shell(fmt.Sprintf("cd %s && git rev-parse --short HEAD", app.LocalPath))
	if err != nil {
		return false
	}
	local := (*out)[0]
	out, err = Shell(fmt.Sprintf("ssh otto@%s readlink -f %s/%s", app.RemoteHost, app.RemotePath, app.CurrentDirName))
	if err != nil {
		return false
	}
	o := (*out)[0]
	rev := o[strings.LastIndex(o, "/")+1:]
	return local == rev
}
