package utils

// AppID is the reverse-DNS identifier passed to Fyne.
const AppID = "io.otto-torino.f8a"

// AppName is the executable and .desktop file name (f8a.desktop). It is also
// used as the X11 WM_CLASS so desktop environments can match the running
// window to its launcher entry, as the desktop entry spec expects.
const AppName = "f8a"

// AppTitle is the main window title and the Name= of the packaged .desktop
// file. The root Makefile reads it with sed: keep the line format
// `const AppTitle = "..."` intact.
const AppTitle = "Otto Frontend Apps Manager"
