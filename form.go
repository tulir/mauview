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

type formItem struct {
	genericChild
}

type Form struct {
	items   []formItem
	focused *formItem
}

func NewForm() *Form {
	return &Form{
		items: []formItem{},
		focused: nil,
	}
}

func (form *Form) Draw() {

}

func (form *Form) NextItem() {

}

func (form *Form) PreviousItem() {

}

func (form *Form) OnKeyEvent(event KeyEvent) bool {
	switch event.Key() {
	case tcell.KeyTab:
		form.NextItem()
	case tcell.KeyBacktab:
		form.PreviousItem()
	}
	if form.focused != nil {
		return form.focused.target.OnKeyEvent(event)
	}
	return false
}

func (form *Form) OnPasteEvent(event PasteEvent) bool {
	if form.focused != nil {
		return form.focused.target.OnPasteEvent(event)
	}
	return false
}

func (form *Form) OnMouseEvent(event MouseEvent) bool {
	return false
}
