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

type TextField struct {
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
	tf.text = text
	return tf
}

func (tf *TextField) SetTextColor(color tcell.Color) *TextField {
	tf.style = tf.style.Foreground(color)
	return tf
}

func (tf *TextField) SetBackgroundColor(color tcell.Color) *TextField {
	tf.style = tf.style.Background(color)
	return tf
}
func (tf *TextField) SetStyle(style tcell.Style) *TextField {
	tf.style = style
	return tf
}

func (tf *TextField) Draw(screen Screen) {
	width, _ := screen.Size()
	screen.SetStyle(tf.style)
	screen.Clear()
	printWithStyle(screen, tf.text, 0, 0, width, AlignLeft, tf.style)
}
