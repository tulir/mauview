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

type Box struct {
	Border      bool
	BorderStyle tcell.Style
	Inner       Component
	innerScreen *ProxyScreen
	focused     bool
}

func NewBox(inner Component) *Box {
	return &Box{
		Border:      true,
		BorderStyle: tcell.StyleDefault,
		Inner:       inner,
		innerScreen: &ProxyScreen{offsetX: 1, offsetY: 1},
	}
}

func (box *Box) Focus() {
	box.focused = true
	focusable, ok := box.Inner.(Focusable)
	if ok {
		focusable.Focus()
	}
}

func (box *Box) Blur() {
	box.focused = false
	focusable, ok := box.Inner.(Focusable)
	if ok {
		focusable.Blur()
	}
}

func (box *Box) Draw(screen Screen) {
	width, height := screen.Size()
	if !box.Border || width < 2 || height < 2 {
		return
	}
	var vertical, horizontal, topLeft, topRight, bottomLeft, bottomRight rune
	if box.focused {
		horizontal = Borders.HorizontalFocus
		vertical = Borders.VerticalFocus
		topLeft = Borders.TopLeftFocus
		topRight = Borders.TopRightFocus
		bottomLeft = Borders.BottomLeftFocus
		bottomRight = Borders.BottomRightFocus
	} else {
		horizontal = Borders.Horizontal
		vertical = Borders.Vertical
		topLeft = Borders.TopLeft
		topRight = Borders.TopRight
		bottomLeft = Borders.BottomLeft
		bottomRight = Borders.BottomRight
	}
	for x := 0; x < width; x++ {
		screen.SetContent(x, 0, horizontal, nil, box.BorderStyle)
		screen.SetContent(x, height-1, horizontal, nil, box.BorderStyle)
	}
	for y := 0; y < height; y++ {
		screen.SetContent(0, y, vertical, nil, box.BorderStyle)
		screen.SetContent(width-1, y, vertical, nil, box.BorderStyle)
	}
	screen.SetContent(0, 0, topLeft, nil, box.BorderStyle)
	screen.SetContent(width-1, 0, topRight, nil, box.BorderStyle)
	screen.SetContent(0, height-1, bottomLeft, nil, box.BorderStyle)
	screen.SetContent(width-1, height-1, bottomRight, nil, box.BorderStyle)

	if box.Inner != nil {
		box.innerScreen.width = width - 2
		box.innerScreen.height = height - 2
		box.innerScreen.parent = screen
		box.Inner.Draw(box.innerScreen)
	}
}

func (box *Box) OnKeyEvent(event KeyEvent) bool {
	if box.Inner != nil {
		return box.Inner.OnKeyEvent(event)
	}
	return false
}

func (box *Box) OnPasteEvent(event PasteEvent) bool {
	if box.Inner != nil {
		return box.Inner.OnPasteEvent(event)
	}
	return false
}

func (box *Box) OnMouseEvent(event MouseEvent) bool {
	if event.Buttons() == tcell.Button1 {
		box.Focus()
	}
	if box.Inner != nil {
		if box.Border {
			event = OffsetMouseEvent(event, -1, -1)
		}
		return box.Inner.OnMouseEvent(event)
	}
	return false
}
