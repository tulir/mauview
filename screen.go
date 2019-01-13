// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"maunium.net/go/tcell"
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
	parent           Screen
	offsetX, offsetY int
	width, height    int
	style            tcell.Style
}

func NewProxyScreen(parent Screen, offsetX, offsetY, width, height int) Screen {
	return &ProxyScreen{
		parent:  parent,
		offsetX: offsetX,
		offsetY: offsetY,
		width:   width,
		height:  height,
		style:   tcell.StyleDefault,
	}
}

func (ss *ProxyScreen) Clear() {
	ss.Fill(' ', ss.style)
}

func (ss *ProxyScreen) Fill(r rune, style tcell.Style) {
	for x := ss.offsetX; x < ss.offsetX+ss.width; x++ {
		for y := ss.offsetY; y < ss.offsetY+ss.height; y++ {
			ss.parent.SetCell(x, y, style, r)
		}
	}
}

func (ss *ProxyScreen) SetStyle(style tcell.Style) {
	ss.style = style
}

func (ss *ProxyScreen) adjustCoordinates(x, y int) (int, int, bool) {
	if x < 0 || y < 0 || (ss.width >= 0 && x >= ss.width) || (ss.height >= 0 && y >= ss.height) {
		return -1, -1, false
	}

	x += ss.offsetX
	y += ss.offsetY
	return x, y, true
}

func (ss *ProxyScreen) SetCell(x, y int, style tcell.Style, ch ...rune) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.parent.SetCell(x, y, style, ch...)
	}
}

func (ss *ProxyScreen) GetContent(x, y int) (mainc rune, combc []rune, style tcell.Style, width int) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		return ss.parent.GetContent(x, y)
	}
	return 0, nil, tcell.StyleDefault, 0
}

func (ss *ProxyScreen) SetContent(x int, y int, mainc rune, combc []rune, style tcell.Style) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.parent.SetContent(x, y, mainc, combc, style)
	}
}

func (ss *ProxyScreen) ShowCursor(x, y int) {
	x, y, ok := ss.adjustCoordinates(x, y)
	if ok {
		ss.parent.ShowCursor(x, y)
	}
}

func (ss *ProxyScreen) HideCursor() {
	ss.parent.HideCursor()
}

// Size returns the size of this subscreen.
//
// If the subscreen doesn't fit in the parent with the set offset and size,
// the returned size is whatever can actually be rendered.
func (ss *ProxyScreen) Size() (width int, height int) {
	width, height = ss.parent.Size()
	width -= ss.offsetX
	height -= ss.offsetY
	if width > ss.width {
		width = ss.width
	}
	if height > ss.height {
		height = ss.height
	}
	return
}

func (ss *ProxyScreen) Colors() int {
	return ss.parent.Colors()
}

func (ss *ProxyScreen) CharacterSet() string {
	return ss.parent.CharacterSet()
}

func (ss *ProxyScreen) CanDisplay(r rune, checkFallbacks bool) bool {
	return ss.parent.CanDisplay(r, checkFallbacks)
}

func (ss *ProxyScreen) HasKey(key tcell.Key) bool {
	return ss.parent.HasKey(key)
}
