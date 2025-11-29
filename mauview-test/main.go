// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"go.mau.fi/mauview"
)

type Text struct {
	tv *mauview.TextView
}

func (text *Text) Draw(screen mauview.Screen) {
	text.tv.Draw(screen)
}

func (text *Text) OnKeyEvent(event mauview.KeyEvent) bool {
	if event.Key() == tcell.KeyCtrlC || (event.Rune() == 'c' && event.Modifiers() == tcell.ModCtrl) {
		go func() {
			app.Stop()
			os.Exit(1)
		}()
	}
	_, _ = fmt.Fprintf(
		text.tv, "Key=%s (%d) Rune=%s (%d) Mod=%s (%d)\n",
		keyName(event.Key()), event.Key(), runeName(event.Rune()), event.Rune(), modName(event.Modifiers()), event.Modifiers())
	text.tv.ScrollToEnd()
	text.tv.OnKeyEvent(event)
	return true
}

func (text *Text) OnPasteEvent(event mauview.PasteEvent) bool {
	return text.tv.OnPasteEvent(event)
}

func (text *Text) OnMouseEvent(event mauview.MouseEvent) bool {
	return text.tv.OnMouseEvent(event)
}

var app *mauview.Application

func main() {
	app = mauview.NewApplication()
	textComp := &Text{mauview.NewTextView()}
	app.SetRoot(mauview.NewBox(textComp))
	err := app.Start()
	if err != nil {
		panic(err)
	}
}

func keyName(key tcell.Key) string {
	name, ok := tcell.KeyNames[key]
	if ok {
		return name
	}
	if key == tcell.KeyRune {
		return "Rune"
	}
	return "Unknown"
}

func runeName(r rune) string {
	switch r {
	case 0:
		return "NUL"
	case '\n':
		return "LF"
	case '\r':
		return "CR"
	case '\t':
		return "TAB"
	case ' ':
		return "Space"
	default:
		if r < 32 {
			return fmt.Sprintf("Ctrl+%c", r+'A'-1)
		}
		return string(r)
	}
}

func modName(mod tcell.ModMask) (name string) {
	if mod&tcell.ModCtrl != 0 {
		name += "Ctrl+"
	}
	if mod&tcell.ModAlt != 0 {
		name += "Alt+"
	}
	if mod&tcell.ModShift != 0 {
		name += "Shift+"
	}
	if mod&tcell.ModMeta != 0 {
		name += "Meta+"
	}
	if name == "" {
		name = "None"
	} else {
		name = name[:len(name)-1]
	}
	return
}
