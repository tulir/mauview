// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"sync"

	"go.mau.fi/tcell"
)

type TextField struct {
	sync.Mutex
	*SimpleEventHandler
	text  string
	style tcell.Style
}

func NewTextField() *TextField {
	return &TextField{
		SimpleEventHandler: &SimpleEventHandler{},

		text:  "",
		style: tcell.StyleDefault.Foreground(Styles.PrimaryTextColor),
	}
}

func (tf *TextField) SetText(text string) *TextField {
	tf.Lock()
	tf.text = text
	tf.Unlock()
	return tf
}

func (tf *TextField) SetTextColor(color tcell.Color) *TextField {
	tf.Lock()
	tf.style = tf.style.Foreground(color)
	tf.Unlock()
	return tf
}

func (tf *TextField) SetBackgroundColor(color tcell.Color) *TextField {
	tf.Lock()
	tf.style = tf.style.Background(color)
	tf.Unlock()
	return tf
}

func (tf *TextField) SetStyle(style tcell.Style) *TextField {
	tf.Lock()
	tf.style = style
	tf.Unlock()
	return tf
}

func (tf *TextField) Draw(screen Screen) {
	tf.Lock()
	width, _ := screen.Size()
	screen.SetStyle(tf.style)
	screen.Clear()
	PrintWithStyle(screen, tf.text, 0, 0, width, AlignLeft, tf.style)
	tf.Unlock()
}
