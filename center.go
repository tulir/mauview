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

type FractionalCenterer struct {
	center         *Centerer
	minWidth       int
	minHeight      int
	fractionWidth  float64
	fractionHeight float64
}

func FractionalCenter(target Component, minWidth, minHeight int, fractionalWidth, fractionalHeight float64) *FractionalCenterer {
	return &FractionalCenterer{
		center:         Center(target, 0, 0),
		minWidth:       minWidth,
		minHeight:      minHeight,
		fractionWidth:  fractionalWidth,
		fractionHeight: fractionalHeight,
	}
}

func (fc *FractionalCenterer) SetAlwaysFocusChild(always bool) *FractionalCenterer {
	fc.center.alwaysFocusChild = always
	return fc
}

func (fc *FractionalCenterer) Blur()  { fc.center.Blur() }
func (fc *FractionalCenterer) Focus() { fc.center.Focus() }

func (fc *FractionalCenterer) OnMouseEvent(evt MouseEvent) bool { return fc.center.OnMouseEvent(evt) }
func (fc *FractionalCenterer) OnKeyEvent(evt KeyEvent) bool     { return fc.center.OnKeyEvent(evt) }
func (fc *FractionalCenterer) OnPasteEvent(evt PasteEvent) bool { return fc.center.OnPasteEvent(evt) }

func (fc *FractionalCenterer) Draw(screen Screen) {
	width, height := screen.Size()
	width = int(float64(width) * fc.fractionWidth)
	height = int(float64(height) * fc.fractionHeight)
	if width < fc.minWidth {
		width = fc.minWidth
	}
	if height < fc.minHeight {
		height = fc.minHeight
	}
	fc.center.SetSize(width, height)
	fc.center.Draw(screen)
}

type Centerer struct {
	target           Component
	screen           *ProxyScreen
	childFocused     bool
	alwaysFocusChild bool
}

func Center(target Component, width, height int) *Centerer {
	return &Centerer{
		target:           target,
		screen:           &ProxyScreen{Style: tcell.StyleDefault, Width: width, Height: height},
		childFocused:     false,
		alwaysFocusChild: false,
	}
}

func (center *Centerer) SetHeight(height int) {
	center.screen.Height = height
}

func (center *Centerer) SetWidth(width int) {
	center.screen.Width = width
}

func (center *Centerer) SetSize(width, height int) {
	center.screen.Width = width
	center.screen.Height = height
}

func (center *Centerer) SetAlwaysFocusChild(always bool) *Centerer {
	center.alwaysFocusChild = always
	return center
}

func (center *Centerer) Draw(screen Screen) {
	totalWidth, totalHeight := screen.Size()
	paddingX := (totalWidth - center.screen.Width) / 2
	paddingY := (totalHeight - center.screen.Height) / 2
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
	if x < 0 || y < 0 || x > center.screen.Width || y > center.screen.Height {
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
