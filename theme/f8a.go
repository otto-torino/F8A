package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type F8aTheme struct{}

var _ fyne.Theme = (*F8aTheme)(nil)

func (t F8aTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (t F8aTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t F8aTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t F8aTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == "text" {
		return 12
	}
	return theme.DefaultTheme().Size(name)
}

func (t F8aTheme) SidebarBg(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.RGBA{R: 245, G: 245, B: 245, A: 255}
	}
	return color.RGBA{R: 33, G: 33, B: 33, A: 255}
}
