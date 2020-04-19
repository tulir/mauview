// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"errors"
	"sync"
	"time"

	"maunium.net/go/tcell"
)

type Component interface {
	Draw(screen Screen)
	OnKeyEvent(event KeyEvent) bool
	OnPasteEvent(event PasteEvent) bool
	OnMouseEvent(event MouseEvent) bool
}

type Focusable interface {
	Focus()
	Blur()
}

type Application struct {
	sync.RWMutex
	screen            tcell.Screen
	Root              Component
	events            chan tcell.Event
	updates           chan func()
	screenReplacement chan tcell.Screen
	redrawTicker      *time.Ticker
}

const queueSize = 255

func NewApplication() *Application {
	return &Application{
		events:            make(chan tcell.Event, queueSize),
		updates:           make(chan func(), queueSize),
		screenReplacement: make(chan tcell.Screen, 1),
		redrawTicker:      time.NewTicker(1 * time.Minute),
	}
}

func (app *Application) receiveNewScreen() (bool, error) {
	screen := <-app.screenReplacement
	if screen == nil {
		app.events <- nil
		return false, nil
	}

	app.Lock()
	app.screen = screen
	app.Unlock()
	if err := screen.Init(); err != nil {
		return true, err
	}
	app.screen.EnableMouse()
	app.Redraw()
	return true, nil
}

func (app *Application) makeNewScreen() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	app.screenReplacement <- screen
	return nil
}

func (app *Application) SetRedrawTicker(tick time.Duration) {
	app.redrawTicker.Stop()
	app.redrawTicker = time.NewTicker(tick)
}

func (app *Application) Start() error {
	if app.Root == nil {
		return errors.New("root component not set")
	}

	err := app.makeNewScreen()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			if app.screen != nil {
				app.screen.Fini()
			}
			panic(p)
		}
	}()

	_, err = app.receiveNewScreen()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				if app.screen != nil {
					app.screen.Fini()
				}
				panic(p)
			}
		}()
		defer wg.Done()
		for {
			app.RLock()
			screen := app.screen
			app.RUnlock()
			if screen == nil {
				break
			}

			event := screen.PollEvent()
			if event == nil {
				ok, err := app.receiveNewScreen()
				if err != nil {
					panic(err)
				} else if !ok {
					break
				}
			} else {
				app.events <- event
			}
		}
	}()

MainLoop:
	for {
		select {
		case event := <-app.events:
			switch event := event.(type) {
			case nil:
				break MainLoop
			case *tcell.EventKey:
				if app.Root.OnKeyEvent(event) {
					app.redraw() // app.update()
				}
			case *tcell.EventPaste:
				if app.Root.OnPasteEvent(event) {
					app.redraw() // app.update()
				}
			case *tcell.EventMouse:
				if app.Root.OnMouseEvent(event) {
					app.redraw() // app.update()
				}
			case *tcell.EventResize:
				app.screen.Clear()
				app.redraw()
			}
		case <-app.redrawTicker.C:
			app.redraw()
		case updater := <-app.updates:
			updater()
		}
	}

	wg.Wait()
	return nil
}

func (app *Application) Stop() {
	app.Lock()
	defer app.Unlock()
	screen := app.screen
	if screen == nil {
		return
	}
	app.screen = nil
	screen.Fini()
	app.screenReplacement <- nil
}

func (app *Application) Suspend(wait func()) bool {
	app.RLock()
	screen := app.screen
	app.RUnlock()
	if screen == nil {
		return false
	}
	screen.Fini()
	wait()
	_ = app.makeNewScreen()
	return true
}

func (app *Application) QueueUpdate(update func()) {
	app.updates <- update
}

// Screen returns the main tcell screen currently used in the app.
func (app *Application) Screen() tcell.Screen {
	return app.screen
}

func (app *Application) Redraw() {
	app.QueueUpdate(app.redraw)
}

func (app *Application) redraw() {
	app.screen.HideCursor()
	app.Root.Draw(app.screen)
	app.update()
}

func (app *Application) Update() {
	app.QueueUpdate(app.update)
}

func (app *Application) update() {
	if app.screen != nil {
		app.screen.Show()
	}
}
