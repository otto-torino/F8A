# Sidebar Professional Styling Design

## Overview

Enhance the sidebar menu visual styling to achieve a modern, professional appearance inspired by contemporary IDEs like VS Code.

## Problem

The current sidebar implementation uses basic styling with minimal visual refinement:
- Plain list widget with default styling
- Limited visual hierarchy
- No hover or interaction feedback
- Lacks polish and professional appearance

## Design Goals

Create a modern, sleek sidebar with:
- Clear visual hierarchy
- Professional color palette and spacing
- Smooth hover and selection interactions
- Theme-aware styling (light/dark modes)

## Visual Design Specification

### Layout Structure

**Sidebar Container:**
- Right edge border (1px, theme-aware)
- Generous padding around content
- Visual separator between title and app list
- Proper spacing hierarchy

**Title Section:**
- "Apps" label in orange accent (#FF9900)
- Size: 18pt (current)
- Increased padding for prominence
- Horizontal separator below

**App List Section:**
- Custom container-based list (not widget.List)
- Individual clickable items with full styling control
- Proper spacing between items
- Scrollable when needed

### Color Palette

**Light Theme:**
- Sidebar background: #F5F5F5
- Border: #E0E0E0
- Separator: #DADADA
- List item hover: #EBEBEB
- List item selected: #FFE6CC
- Text: Dark gray (theme default)

**Dark Theme:**
- Sidebar background: #212121
- Border: #3A3A3A
- Separator: #404040
- List item hover: #2C2C2C
- List item selected: #4A3A2A
- Text: Light gray/white (theme default)

### Interactive States

**List Items:**
- Default: Transparent background, standard text
- Hover: Background color change, smooth transition
- Selected: Distinct background color, maintained during hover
- Active/Pressed: Subtle feedback (slight darkening)

### Spacing

- Sidebar padding: 12-16px
- Title bottom margin: 12px
- Separator margin: 8px vertical
- List item padding: 10px vertical, 12px horizontal
- List item gap: 4px

## Implementation Approach

### Option A: Custom List Items (Selected)

Replace `widget.List` with custom implementation:

1. Create container with `container.NewVBox()`
2. Build individual list items as custom widgets or buttons
3. Each item is a clickable container with:
   - Custom background rendering
   - Hover state tracking
   - Selection state styling
   - Click handler for app selection

**Advantages:**
- Full styling control
- Smooth hover/selection effects
- Professional appearance
- Better spacing control

**Implementation Details:**
- Create custom `SidebarItem` widget or use styled buttons
- Track hover state for visual feedback
- Track selected app ID for selection state
- Use canvas.Rectangle for backgrounds
- Apply smooth color transitions

### Theme Integration

Extend `theme/f8a.go` with new color methods:
- `SidebarBorder(variant)` - border colors
- `SidebarSeparator(variant)` - separator colors
- `SidebarItemHover(variant)` - hover state
- `SidebarItemSelected(variant)` - selected state

### Component Structure

```
Sidebar Container (w/ background + border)
├── Padding Container
    ├── Title Section
    │   ├── "Apps" Text
    │   └── Separator Line
    └── App List Container (VBox, scrollable)
        ├── App Item 1 (custom widget)
        ├── App Item 2
        └── ...
```

## Files to Modify

- `components/sidebar.go` - Main implementation
- `theme/f8a.go` - Add new color methods
- Possibly create `components/sidebar_item.go` for custom list item widget

## Success Criteria

- Sidebar has clear visual definition with borders
- Title section has proper hierarchy and spacing
- App items have visible hover effects
- Selected app is clearly indicated
- Smooth transitions between states
- Works correctly in both light and dark themes
- Maintains all existing functionality (selection, updates, etc.)
