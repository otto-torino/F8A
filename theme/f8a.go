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

func (t F8aTheme) SidebarBorder(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.RGBA{R: 224, G: 224, B: 224, A: 255} // #E0E0E0
	}
	return color.RGBA{R: 58, G: 58, B: 58, A: 255} // #3A3A3A
}

func (t F8aTheme) SidebarSeparator(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.RGBA{R: 218, G: 218, B: 218, A: 255} // #DADADA
	}
	return color.RGBA{R: 64, G: 64, B: 64, A: 255} // #404040
}

func (t F8aTheme) SidebarItemHover(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.RGBA{R: 235, G: 235, B: 235, A: 255} // #EBEBEB
	}
	return color.RGBA{R: 44, G: 44, B: 44, A: 255} // #2C2C2C
}

func (t F8aTheme) SidebarItemSelected(variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantLight {
		return color.RGBA{R: 255, G: 230, B: 204, A: 255} // #FFE6CC
	}
	return color.RGBA{R: 74, G: 58, B: 42, A: 255} // #4A3A2A
}
