package utils

import (
	"fmt"
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
