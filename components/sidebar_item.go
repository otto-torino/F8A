package components

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/otto-torino/f8a/theme"
)

var _ desktop.Hoverable = (*SidebarItem)(nil)
var _ fyne.Tappable = (*SidebarItem)(nil)

type SidebarItem struct {
	widget.BaseWidget
	text       string
	isSelected bool
	isHovered  bool
	onTapped   func()
	background *canvas.Rectangle
	label      *canvas.Text
	themeVariant fyne.ThemeVariant
}

func NewSidebarItem(text string, onTapped func(), themeVariant fyne.ThemeVariant) *SidebarItem {
	item := &SidebarItem{
		text:         text,
		onTapped:     onTapped,
		themeVariant: themeVariant,
	}
	item.ExtendBaseWidget(item)
	return item
}

func (s *SidebarItem) CreateRenderer() fyne.WidgetRenderer {
	s.background = canvas.NewRectangle(color.Transparent)
	s.background.CornerRadius = 4 // Add subtle rounding

	s.label = canvas.NewText(s.text, color.Black)
	s.label.TextSize = 13

	s.updateColors()

	// Add horizontal and vertical padding
	content := container.NewPadded(s.label)
	objects := []fyne.CanvasObject{s.background, content}

	return widget.NewSimpleRenderer(container.NewStack(objects...))
}

func (s *SidebarItem) SetSelected(selected bool) {
	s.isSelected = selected
	s.updateColors()
	s.Refresh()
}

func (s *SidebarItem) updateColors() {
	t := theme.F8aTheme{}

	if s.isSelected {
		s.background.FillColor = t.SidebarItemSelected(s.themeVariant)
	} else if s.isHovered {
		s.background.FillColor = t.SidebarItemHover(s.themeVariant)
	} else {
		s.background.FillColor = color.Transparent
	}

	if s.themeVariant == 1 { // theme.VariantLight
		s.label.Color = color.RGBA{R: 33, G: 33, B: 33, A: 255}
	} else {
		s.label.Color = color.RGBA{R: 220, G: 220, B: 220, A: 255}
	}
}

func (s *SidebarItem) Tapped(*fyne.PointEvent) {
	if s.onTapped != nil {
		s.onTapped()
	}
}

func (s *SidebarItem) MouseIn(*desktop.MouseEvent) {
	s.isHovered = true
	s.updateColors()
	s.Refresh()
}

func (s *SidebarItem) MouseOut() {
	s.isHovered = false
	s.updateColors()
	s.Refresh()
}

func (s *SidebarItem) MouseMoved(*desktop.MouseEvent) {
	// No-op: hover state is handled by MouseIn/MouseOut
}

func (s *SidebarItem) MinSize() fyne.Size {
	return fyne.NewSize(150, 40)
}
