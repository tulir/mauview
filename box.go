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
	OnKey   func(event *tcell.EventKey) bool
	OnPaste func(event *tcell.EventPaste) bool
	OnMouse func(event *tcell.EventMouse) bool
}

func (box *Box) Draw(screen Screen) {

}

func (box *Box) OnKeyEvent(event *tcell.EventKey) bool {
	if box.OnKey != nil {
		return box.OnKey(event)
	}
	return false
}

func (box *Box) OnPasteEvent(event *tcell.EventPaste) bool {
	if box.OnPaste != nil {
		return box.OnPaste(event)
	}
	return false
}

func (box *Box) OnMouseEvent(event *tcell.EventMouse) bool {
	if box.OnMouse != nil {
		return box.OnMouse(event)
	}
	return false
}
