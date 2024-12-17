package utils

import (
	"sync"

	"fyne.io/fyne/v2"
)

type AppRegistry struct {
	Application *fyne.App
	Window      *fyne.Window
}

var (
	registry *AppRegistry
	once     sync.Once
)

// Get the global registry
func Registry() *AppRegistry {
	if registry == nil {
		once.Do(func() {
			registry = &AppRegistry{}
		})
	}

	return registry
}
