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

// KeyEvent is an interface of the *tcell.EventKey type.
type KeyEvent interface {
	tcell.Event
	// The rune corresponding to the key that was pressed.
	Rune() rune
	// The keyboard key that was pressed.
	Key() tcell.Key
	// The keyboard modifiers that were pressed during the event.
	Modifiers() tcell.ModMask
}

type customPasteEvent struct {
	*tcell.EventPaste
	text string
}

func (cpe customPasteEvent) Text() string {
	return cpe.text
}

// PasteEvent is an interface of the customPasteEvent type.
type PasteEvent interface {
	tcell.Event
	// The text pasted.
	Text() string
}

// MouseEvent is an interface of the *tcell.EventMouse type.
type MouseEvent interface {
	tcell.Event
	// The mouse buttons that were pressed.
	Buttons() tcell.ButtonMask
	// The keyboard modifiers that were pressed during the event.
	Modifiers() tcell.ModMask
	// The current position of the mouse.
	Position() (int, int)
	// Whether or not the event is a mouse move event.
	HasMotion() bool
}

// SimpleEventHandler is a simple implementation of the event handling methods required for components.
type SimpleEventHandler struct {
	OnKey   func(event KeyEvent) bool
	OnPaste func(event PasteEvent) bool
	OnMouse func(event MouseEvent) bool
}

func (seh *SimpleEventHandler) OnKeyEvent(event KeyEvent) bool {
	if seh.OnKey != nil {
		return seh.OnKey(event)
	}
	return false
}

func (seh *SimpleEventHandler) OnPasteEvent(event PasteEvent) bool {
	if seh.OnPaste != nil {
		return seh.OnPaste(event)
	}
	return false
}

func (seh *SimpleEventHandler) OnMouseEvent(event MouseEvent) bool {
	if seh.OnMouse != nil {
		return seh.OnMouse(event)
	}
	return false
}

type NoopEventHandler struct{}

func (neh NoopEventHandler) OnKeyEvent(event KeyEvent) bool {
	return false
}

func (neh NoopEventHandler) OnPasteEvent(event PasteEvent) bool {
	return false
}

func (neh NoopEventHandler) OnMouseEvent(event MouseEvent) bool {
	return false
}

type proxyEventMouse struct {
	MouseEvent
	x int
	y int
}

func (evt *proxyEventMouse) Position() (int, int) {
	return evt.x, evt.y
}

// OffsetMouseEvent creates a new MouseEvent with the given offset.
func OffsetMouseEvent(evt MouseEvent, offsetX, offsetY int) *proxyEventMouse {
	x, y := evt.Position()
	proxy, ok := evt.(*proxyEventMouse)
	if ok {
		evt = proxy.MouseEvent
	}
	return &proxyEventMouse{evt, x + offsetX, y + offsetY}
}
