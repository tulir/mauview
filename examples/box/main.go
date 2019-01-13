// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"maunium.net/go/mauview"
	"maunium.net/go/tcell"
)

type Text struct {
	mauview.Box
	Text string
}

func (text *Text) Draw(screen mauview.Screen) {
	for i, char := range text.Text {
		screen.SetCell(i, 0, tcell.StyleDefault, char)
	}
}

func main() {
	app := mauview.NewApplication()
	grid := mauview.NewGrid(3, 3)
	textComp := &Text{mauview.Box{}, "Hello, World!"}
	textComp.OnKey = func(event *tcell.EventKey) bool {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
		}
		return false
	}
	grid.AddComponent(textComp, 1, 1, 1, 1)
	app.Root = grid
	err := app.Start()
	if err != nil {
		panic(err)
	}
}
