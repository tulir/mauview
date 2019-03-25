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

type Centerer struct {
	target           Component
	screen           *ProxyScreen
	Width            int
	Height           int
	childFocused     bool
	alwaysFocusChild bool
}

func Center(target Component, width, height int) *Centerer {
	return &Centerer{
		target:           target,
		screen:           &ProxyScreen{Style: tcell.StyleDefault, Width: width, Height: height},
		Width:            width,
		Height:           height,
		childFocused:     false,
		alwaysFocusChild: false,
	}
}

func (center *Centerer) SetAlwaysFocusChild(always bool) *Centerer {
	center.alwaysFocusChild = always
	return center
}

func (center *Centerer) Draw(screen Screen) {
	totalWidth, totalHeight := screen.Size()
	paddingX := (totalWidth - center.Width) / 2
	paddingY := (totalHeight - center.Height) / 2
	if paddingX >= 0 {
		center.screen.OffsetX = paddingX
	}
	if paddingY >= 0 {
		center.screen.OffsetY = paddingY
	}
	center.screen.Parent = screen
	center.target.Draw(center.screen)
}

func (center *Centerer) OnKeyEvent(evt KeyEvent) bool {
	return center.target.OnKeyEvent(evt)
}

func (center *Centerer) Focus() {
	if center.alwaysFocusChild {
		center.childFocused = true
		focusable, ok := center.target.(Focusable)
		if ok {
			focusable.Focus()
		}
	}
}

func (center *Centerer) Blur() {
	center.childFocused = false
	focusable, ok := center.target.(Focusable)
	if ok {
		focusable.Blur()
	}
}

func (center *Centerer) OnMouseEvent(evt MouseEvent) bool {
	x, y := evt.Position()
	x -= center.screen.OffsetX
	y -= center.screen.OffsetY
	focusable, ok := center.target.(Focusable)
	if x < 0 || y < 0 || x > center.Width || y > center.Height {
		if ok && evt.Buttons() == tcell.Button1 && !evt.HasMotion() {
			if center.alwaysFocusChild && !center.childFocused {
				focusable.Focus()
				center.childFocused = true
			} else if !center.alwaysFocusChild && center.childFocused {
				center.Blur()
			}
			return true
		}
		return false
	}
	focusChanged := false
	if ok && !center.childFocused && evt.Buttons() == tcell.Button1 && !evt.HasMotion() {
		focusable.Focus()
		center.childFocused = true
		focusChanged = true
	}
	return center.target.OnMouseEvent(OffsetMouseEvent(evt, -center.screen.OffsetX, -center.screen.OffsetY)) || focusChanged
}

func (center *Centerer) OnPasteEvent(evt PasteEvent) bool {
	return center.target.OnPasteEvent(evt)
}
