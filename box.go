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

type KeyCaptureFunc func(event KeyEvent) KeyEvent
type MouseCaptureFunc func(event MouseEvent) MouseEvent
type PasteCaptureFunc func(event PasteEvent) PasteEvent

type Box struct {
	border          bool
	borderStyle     tcell.Style
	backgroundColor *tcell.Color
	keyCapture      KeyCaptureFunc
	mouseCapture    MouseCaptureFunc
	pasteCapture    PasteCaptureFunc
	focusCapture    func() bool
	blurCapture     func() bool
	title           string
	inner           Component
	innerScreen     *ProxyScreen
	focused         bool
}

func NewBox(inner Component) *Box {
	return &Box{
		border:          true,
		borderStyle:     tcell.StyleDefault,
		backgroundColor: &Styles.PrimitiveBackgroundColor,
		inner:           inner,
		innerScreen:     &ProxyScreen{OffsetX: 1, OffsetY: 1},
	}
}

func (box *Box) Focus() {
	box.focused = true
	if box.focusCapture != nil {
		if box.focusCapture() {
			return
		}
	}
	focusable, ok := box.inner.(Focusable)
	if ok {
		focusable.Focus()
	}
}

func (box *Box) Blur() {
	box.focused = false
	if box.blurCapture != nil {
		if box.blurCapture() {
			return
		}
	}
	focusable, ok := box.inner.(Focusable)
	if ok {
		focusable.Blur()
	}
}

func (box *Box) SetBorder(border bool) *Box {
	box.border = border
	if border {
		box.innerScreen.OffsetY = 1
		box.innerScreen.OffsetX = 1
	} else {
		box.innerScreen.OffsetY = 0
		box.innerScreen.OffsetX = 0
	}
	return box
}

func (box *Box) SetBorderStyle(borderStyle tcell.Style) *Box {
	box.borderStyle = borderStyle
	return box
}

func (box *Box) SetTitle(title string) *Box {
	box.title = title
	return box
}

func (box *Box) SetInnerComponent(component Component) *Box {
	box.inner = component
	return box
}

func (box *Box) SetMouseCaptureFunc(mouseCapture MouseCaptureFunc) *Box {
	box.mouseCapture = mouseCapture
	return box
}

func (box *Box) SetKeyCaptureFunc(keyCapture KeyCaptureFunc) *Box {
	box.keyCapture = keyCapture
	return box
}

func (box *Box) SetPasteCaptureFunc(pasteCapture PasteCaptureFunc) *Box {
	box.pasteCapture = pasteCapture
	return box
}

func (box *Box) SetFocusCaptureFunc(focusCapture func() bool) *Box {
	box.focusCapture = focusCapture
	return box
}

func (box *Box) SetBlurCaptureFunc(blurCapture func() bool) *Box {
	box.blurCapture = blurCapture
	return box
}

func (box *Box) SetBackgroundColor(color tcell.Color) *Box {
	box.backgroundColor = &color
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
	borderStyle := box.borderStyle
	if box.backgroundColor != nil {
		borderStyle = borderStyle.Background(*box.backgroundColor)
	}
	for x := 0; x < width; x++ {
		screen.SetContent(x, 0, horizontal, nil, borderStyle)
		screen.SetContent(x, height-1, horizontal, nil, borderStyle)
	}
	Print(screen, box.title, 1, 0, width-2, AlignCenter, Styles.BorderColor)
	for y := 0; y < height; y++ {
		screen.SetContent(0, y, vertical, nil, borderStyle)
		screen.SetContent(width-1, y, vertical, nil, borderStyle)
	}
	screen.SetContent(0, 0, topLeft, nil, borderStyle)
	screen.SetContent(width-1, 0, topRight, nil, borderStyle)
	screen.SetContent(0, height-1, bottomLeft, nil, borderStyle)
	screen.SetContent(width-1, height-1, bottomRight, nil, borderStyle)
}

func (box *Box) Draw(screen Screen) {
	width, height := screen.Size()
	border := false
	if box.backgroundColor != nil {
		screen.SetStyle(tcell.StyleDefault.Background(*box.backgroundColor))
		screen.Clear()
	}
	if box.border && width >= 2 && height >= 2 {
		border = true
		box.drawBorder(screen)
	}

	if box.inner != nil {
		if border {
			box.innerScreen.Width = width - 2
			box.innerScreen.Height = height - 2
		} else {
			box.innerScreen.Width = width
			box.innerScreen.Height = height
		}
		box.innerScreen.Parent = screen
		box.inner.Draw(box.innerScreen)
	}
}

func (box *Box) OnKeyEvent(event KeyEvent) bool {
	if box.keyCapture != nil {
		event = box.keyCapture(event)
		if event == nil {
			return true
		}
	}
	if box.inner != nil {
		return box.inner.OnKeyEvent(event)
	}
	return false
}

func (box *Box) OnPasteEvent(event PasteEvent) bool {
	if box.pasteCapture != nil {
		event = box.pasteCapture(event)
		if event == nil {
			return true
		}
	}
	if box.inner != nil {
		return box.inner.OnPasteEvent(event)
	}
	return false
}

func (box *Box) OnMouseEvent(event MouseEvent) bool {
	if box.border {
		event = OffsetMouseEvent(event, -1, -1)
	}
	x, y := event.Position()
	if x < 0 || y < 0 || x > box.innerScreen.Width || y > box.innerScreen.Height {
		return false
	}
	if box.mouseCapture != nil {
		event = box.mouseCapture(event)
		if event == nil {
			return true
		}
	}
	if box.inner != nil {
		return box.inner.OnMouseEvent(event)
	}
	return false
}
