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

type Form struct {
	*Grid
	items []*gridChild
}

type FormItem interface {
	Component
	Submit(event KeyEvent) bool
}

func NewForm() *Form {
	return &Form{
		Grid: NewGrid(),
	}
}

func (form *Form) Draw(screen Screen) {
	form.Grid.Draw(screen)
}

func (form *Form) FocusNextItem() {
	for i := 0; i < len(form.items)-1; i++ {
		if form.focused == form.items[i] {
			form.setFocused(form.items[i+1])
			return
		}
	}
	form.setFocused(form.items[0])
}

func (form *Form) FocusPreviousItem() {
	for i := len(form.items) - 1; i > 0; i-- {
		if form.focused == form.items[i] {
			form.setFocused(form.items[i-1])
			return
		}
	}
	form.setFocused(form.items[len(form.items)-1])
}

func (form *Form) AddFormItem(comp Component, x, y, width, height int) *Form {
	child := form.Grid.createChild(comp, x, y, width, height)
	form.items = append(form.items, child)
	form.Grid.addChild(child)
	return form
}

func (form *Form) RemoveFormItem(comp Component) *Form {
	for index := len(form.items) - 1; index >= 0; index-- {
		if form.items[index].target == comp {
			form.items = append(form.items[:index], form.items[index+1:]...)
		}
	}
	form.Grid.RemoveComponent(comp)
	return form
}

func (form *Form) OnKeyEvent(event KeyEvent) bool {
	switch event.Key() {
	case tcell.KeyTab:
		form.FocusNextItem()
		return true
	case tcell.KeyBacktab:
		form.FocusPreviousItem()
		return true
	case tcell.KeyEnter:
		if form.focused != nil {
			if fi, ok := form.focused.target.(FormItem); ok {
				if fi.Submit(event) {
					form.FocusNextItem()
					return true
				} else {
					return false
				}
			}
		}
	}
	return form.Grid.OnKeyEvent(event)
}
