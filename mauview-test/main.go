// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"go.mau.fi/tcell"

	"go.mau.fi/mauview"
)

type Text struct {
	mauview.SimpleEventHandler
	Text string
}

func (text *Text) Draw(screen mauview.Screen) {
	for i, char := range text.Text {
		screen.SetCell(i, 0, tcell.StyleDefault, char)
	}
}

func main() {
	app := mauview.NewApplication()
	grid := mauview.NewGrid()
	textComp := &Text{mauview.SimpleEventHandler{}, "Hello, World!"}
	textComp.OnKey = func(event mauview.KeyEvent) bool {
		if event.Key() == tcell.KeyCtrlC || event.Rune() == 'q' {
			app.Stop()
		}
		return false
	}
	grid.SetColumn(0, 25)
	grid.SetRow(1, 15)
	grid.SetRow(3, 5)
	grid.SetRow(4, 3)
	grid.AddComponent(mauview.NewBox(textComp), 1, 0, 2, 2)
	grid.AddComponent(mauview.NewBox(mauview.NewFlex().SetDirection(mauview.FlexRow).
		AddFixedComponent(mauview.NewBox(nil), 10).
		AddProportionalComponent(mauview.NewBox(nil), 3).
		AddProportionalComponent(mauview.NewBox(nil), 1).
		AddFixedComponent(mauview.NewBox(nil), 10)), 0, 0, 1, 3)
	grid.AddComponent(mauview.NewBox(
		mauview.NewGrid().
			AddComponent(&Text{mauview.SimpleEventHandler{}, "Hello, World! (again)"}, 0, 1, 1, 1).
			AddComponent(mauview.NewBox(mauview.NewInputArea().SetPlaceholder("I'm holding a place!")), 0, 0, 2, 1).
			AddComponent(mauview.NewBox(nil), 1, 1, 1, 1)),
		1, 2, 1, 1)
	grid.AddComponent(mauview.NewBox(mauview.Center(mauview.NewBox(nil), 10, 5).SetAlwaysFocusChild(true)), 2, 2, 1, 1)
	grid.AddComponent(mauview.NewBox(nil), 0, 4, 2, 1)
	grid.AddComponent(mauview.NewBox(mauview.NewInputField()), 0, 3, 3, 1)
	app.SetRoot(mauview.NewBox(grid))
	err := app.Start()
	if err != nil {
		panic(err)
	}
}
