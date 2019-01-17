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
	border      bool
	borderStyle tcell.Style
	inner       Component
	innerScreen *ProxyScreen
	focused     bool
}

func NewBox(inner Component) *Box {
	return &Box{
		border:      true,
		borderStyle: tcell.StyleDefault,
		inner:       inner,
		innerScreen: &ProxyScreen{offsetX: 1, offsetY: 1},
	}
}

func (box *Box) Focus() {
	box.focused = true
	focusable, ok := box.inner.(Focusable)
	if ok {
		focusable.Focus()
	}
}

func (box *Box) Blur() {
	box.focused = false
	focusable, ok := box.inner.(Focusable)
	if ok {
		focusable.Blur()
	}
}

func (box *Box) SetBorder(border bool) *Box {
	box.border = border
	return box
}

func (box *Box) SetBorderStyle(borderStyle tcell.Style) *Box {
	box.borderStyle = borderStyle
	return box
}

func (box *Box) SetInnerComponent(component Component) *Box {
	box.inner = component
	return box
}

func (box *Box) drawBorder(screen Screen) {
	width, height := screen.Size()
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
		screen.SetContent(x, 0, horizontal, nil, box.borderStyle)
		screen.SetContent(x, height-1, horizontal, nil, box.borderStyle)
	}
	for y := 0; y < height; y++ {
		screen.SetContent(0, y, vertical, nil, box.borderStyle)
		screen.SetContent(width-1, y, vertical, nil, box.borderStyle)
	}
	screen.SetContent(0, 0, topLeft, nil, box.borderStyle)
	screen.SetContent(width-1, 0, topRight, nil, box.borderStyle)
	screen.SetContent(0, height-1, bottomLeft, nil, box.borderStyle)
	screen.SetContent(width-1, height-1, bottomRight, nil, box.borderStyle)
}

func (box *Box) Draw(screen Screen) {
	width, height := screen.Size()
	border := false
	if box.border && width >= 2 && height >= 2 {
		border = true
		box.drawBorder(screen)
	}

	if box.inner != nil {
		if border {
			width -= 2
			height -= 2
		}
		box.innerScreen.width = width
		box.innerScreen.height = height
		box.innerScreen.parent = screen
		box.inner.Draw(box.innerScreen)
	}
}

func (box *Box) OnKeyEvent(event KeyEvent) bool {
	if box.inner != nil {
		return box.inner.OnKeyEvent(event)
	}
	return false
}

func (box *Box) OnPasteEvent(event PasteEvent) bool {
	if box.inner != nil {
		return box.inner.OnPasteEvent(event)
	}
	return false
}

func (box *Box) OnMouseEvent(event MouseEvent) bool {
	if event.Buttons() == tcell.Button1 {
		box.Focus()
	}
	if box.inner != nil {
		if box.border {
			event = OffsetMouseEvent(event, -1, -1)
		}
		x, y := event.Position()
		if x < 0 || y < 0 || x > box.innerScreen.width || y > box.innerScreen.height {
			return false
		}
		return box.inner.OnMouseEvent(event)
	}
	return false
}
