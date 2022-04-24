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

type Button struct {
	text         string
	style        tcell.Style
	focusedStyle tcell.Style
	focused      bool
	onClick      func()
}

func NewButton(text string) *Button {
	return &Button{
		text:         text,
		style:        tcell.StyleDefault.Background(Styles.ContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
		focusedStyle: tcell.StyleDefault.Background(Styles.MoreContrastBackgroundColor).Foreground(Styles.PrimaryTextColor),
	}
}

func (b *Button) SetText(text string) *Button {
	b.text = text
	return b
}

func (b *Button) SetForegroundColor(color tcell.Color) *Button {
	b.style = b.style.Foreground(color)
	return b
}

func (b *Button) SetBackgroundColor(color tcell.Color) *Button {
	b.style = b.style.Background(color)
	return b
}

func (b *Button) SetFocusedForegroundColor(color tcell.Color) *Button {
	b.focusedStyle = b.focusedStyle.Foreground(color)
	return b
}

func (b *Button) SetFocusedBackgroundColor(color tcell.Color) *Button {
	b.focusedStyle = b.focusedStyle.Background(color)
	return b
}

func (b *Button) SetStyle(style tcell.Style) *Button {
	b.style = style
	return b
}

func (b *Button) SetFocusedStyle(style tcell.Style) *Button {
	b.focusedStyle = style
	return b
}

func (b *Button) SetOnClick(fn func()) *Button {
	b.onClick = fn
	return b
}

func (b *Button) Focus() {
	b.focused = true
}

func (b *Button) Blur() {
	b.focused = false
}

func (b *Button) Draw(screen Screen) {
	width, _ := screen.Size()
	style := b.style
	if b.focused {
		style = b.focusedStyle
	}
	screen.SetStyle(style)
	PrintWithStyle(screen, b.text, 0, 0, width, AlignCenter, style)
}

func (b *Button) Submit(event KeyEvent) bool {
	if b.onClick != nil {
		b.onClick()
	}
	return true
}

func (b *Button) OnKeyEvent(event KeyEvent) bool {
	if event.Key() == tcell.KeyEnter {
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}
	return false
}

func (b *Button) OnMouseEvent(event MouseEvent) bool {
	if event.Buttons() == tcell.Button1 && !event.HasMotion() {
		if b.onClick != nil {
			b.onClick()
		}
		return true
	}
	return false
}

func (b *Button) OnPasteEvent(event PasteEvent) bool {
	return false
}
