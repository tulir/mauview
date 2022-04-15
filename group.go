// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

type genericChild struct {
	screen *ProxyScreen
	target Component
}

func (child genericChild) Focus() {
	focusable, ok := child.target.(Focusable)
	if ok {
		focusable.Focus()
	}
}

func (child genericChild) Blur() {
	focusable, ok := child.target.(Focusable)
	if ok {
		focusable.Blur()
	}
}
