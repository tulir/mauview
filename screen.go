// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"go.mau.fi/tcell"
)

// Screen is a subset of the tcell Screen.
// See https://godoc.org/maunium.net/go/tcell#Screen for documentation.
type Screen interface {
	Clear()
	Fill(rune, tcell.Style)
	SetStyle(style tcell.Style)
	SetCell(x, y int, style tcell.Style, ch ...rune)
	GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int)
	SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style)
	ShowCursor(x int, y int)
	HideCursor()
	Size() (int, int)
	Colors() int
	CharacterSet() string
	CanDisplay(r rune, checkFallbacks bool) bool
	HasKey(tcell.Key) bool
}

// ProxyScreen is a proxy to a tcell Screen with a specific allowed drawing area.
type ProxyScreen struct {
	Parent           Screen
	OffsetX, OffsetY int
	Width, Height    int
	Style            tcell.Style
}

func NewProxyScreen(parent Screen, offsetX, offsetY, width, height int) Screen {
	return &ProxyScreen{
		Parent:  parent,
		OffsetX: offsetX,
		OffsetY: offsetY,
		Width:   width,
		Height:  height,
		Style:   tcell.StyleDefault,
	}
}

func (ss *ProxyScreen) IsInArea(x, y int) bool {
	return x >= ss.OffsetX && x <= ss.OffsetX+ss.Width &&
		y >= ss.OffsetY && y <= ss.OffsetY+ss.Height
}

func (ss *ProxyScreen) YEnd() int {
	return ss.OffsetY + ss.Height
}

func (ss *ProxyScreen) XEnd() int {
	return ss.OffsetX + ss.Width
}

func (ss *ProxyScreen) OffsetMouseEvent(event MouseEvent) MouseEvent {
	return OffsetMouseEvent(event, -ss.OffsetX, -ss.OffsetY)
}

func (ss *ProxyScreen) Clear() {
	ss.Fill(' ', ss.Style)
}

func (ss *ProxyScreen) Fill(r rune, style tcell.Style) {
	for x := ss.OffsetX; x < ss.XEnd(); x++ {
		for y := ss.OffsetY; y < ss.YEnd(); y++ {
			ss.Parent.SetCell(x, y, style, r)
		}
	}
}

func (ss *ProxyScreen) SetStyle(style tcell.Style) {
	ss.Style = style
}

func (ss *ProxyScreen) adjustCoordinates(x, y int) (int, int, bool) {
	if x < 0 || y < 0 || (ss.Width >= 0 && x >= ss.Width) || (ss.Height >= 0 && y >= ss.Height) {
		return -1, -1, false
	}

	x += ss.OffsetX
	y += ss.OffsetY
	return x, y, true
}

func (ss *ProxyScreen) SetCell(x, y int, style tcell.Style, ch ...rune) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.Parent.SetCell(x, y, style, ch...)
	}
}

func (ss *ProxyScreen) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		return ss.Parent.GetContent(x, y)
	}
	return 0, nil, tcell.StyleDefault, 0
}

func (ss *ProxyScreen) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.Parent.SetContent(x, y, mainc, combc, style)
	}
}

func (ss *ProxyScreen) ShowCursor(x, y int) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.Parent.ShowCursor(x, y)
	}
}

func (ss *ProxyScreen) HideCursor() {
	ss.Parent.HideCursor()
}

// Size returns the size of this subscreen.
//
// If the subscreen doesn't fit in the parent with the set offset and size,
// the returned size is whatever can actually be rendered.
func (ss *ProxyScreen) Size() (width int, height int) {
	width, height = ss.Parent.Size()
	width -= ss.OffsetX
	height -= ss.OffsetY
	if width > ss.Width {
		width = ss.Width
	}
	if height > ss.Height {
		height = ss.Height
	}
	return
}

func (ss *ProxyScreen) Colors() int {
	return ss.Parent.Colors()
}

func (ss *ProxyScreen) CharacterSet() string {
	return ss.Parent.CharacterSet()
}

func (ss *ProxyScreen) CanDisplay(r rune, checkFallbacks bool) bool {
	return ss.Parent.CanDisplay(r, checkFallbacks)
}

func (ss *ProxyScreen) HasKey(key tcell.Key) bool {
	return ss.Parent.HasKey(key)
}
