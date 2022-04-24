// mauview - A Go TUI library based on tcell.
// Copyright © 2022 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.mau.fi/tcell"
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

type FocusableComponent interface {
	Component
	Focusable
}

type Application struct {
	screenLock   sync.RWMutex
	screen       tcell.Screen
	root         Component
	updates      chan interface{}
	redrawTicker *time.Ticker
	stop         chan struct{}
	waitForStop  chan struct{}
	alwaysClear  bool
}

const queueSize = 255

func NewApplication() *Application {
	return &Application{
		updates:      make(chan interface{}, queueSize),
		redrawTicker: time.NewTicker(1 * time.Minute),
		stop:         make(chan struct{}, 1),
		alwaysClear:  true,
	}
}

func newScreen(events chan tcell.Event) (tcell.Screen, error) {
	if screen, err := tcell.NewScreen(); err != nil {
		return nil, fmt.Errorf("failed to create screen: %w", err)
	} else if err = screen.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize screen: %w", err)
	} else {
		screen.EnableMouse()
		screen.EnablePaste()
		go screen.ChannelEvents(events, nil)
		return screen, nil
	}
}

func (app *Application) SetRedrawTicker(tick time.Duration) {
	app.redrawTicker.Stop()
	app.redrawTicker = time.NewTicker(tick)
}

func (app *Application) Start() error {
	if app.root == nil {
		return errors.New("root component not set")
	}

	events := make(chan tcell.Event, queueSize)
	screen, err := newScreen(events)
	if err != nil {
		return err
	}

	app.screenLock.Lock()
	app.screen = screen
	app.screenLock.Unlock()
	app.waitForStop = make(chan struct{})

	defer func() {
		app.screenLock.Lock()
		app.screen = nil
		app.screenLock.Unlock()
		close(app.waitForStop)
		if screen != nil {
			screen.Fini()
		}
	}()

	var pasteBuffer strings.Builder
	var isPasting bool

	for {
		var redraw bool
		var clear bool
		select {
		case eventInterface := <-events:
			switch event := eventInterface.(type) {
			case *tcell.EventKey:
				if isPasting {
					switch event.Key() {
					case tcell.KeyRune:
						pasteBuffer.WriteRune(event.Rune())
					case tcell.KeyEnter:
						pasteBuffer.WriteByte('\n')
					}
				} else {
					redraw = app.root.OnKeyEvent(event)
				}
			case *tcell.EventPaste:
				if event.Start() {
					isPasting = true
					pasteBuffer.Reset()
				} else {
					customEvt := customPasteEvent{event, pasteBuffer.String()}
					isPasting = false
					pasteBuffer.Reset()
					redraw = app.root.OnPasteEvent(customEvt)
				}
			case *tcell.EventMouse:
				redraw = app.root.OnMouseEvent(event)
			case *tcell.EventResize:
				clear = true
				redraw = true
			}
		case <-app.redrawTicker.C:
			redraw = true
		case updaterInterface := <-app.updates:
			switch updater := updaterInterface.(type) {
			case redrawUpdate:
				redraw = true
			case setRootUpdate:
				app.root = updater.newRoot
				focusable, ok := app.root.(Focusable)
				if ok {
					focusable.Focus()
				}
				redraw = true
				clear = true
			case suspendUpdate:
				err = screen.Suspend()
				if err != nil {
					// This shouldn't fail
					panic(err)
				}
				updater.wait()
				err = screen.Resume()
				if err != nil {
					screen.Fini()
					fmt.Println("Failed to resume screen:", err)
					os.Exit(40)
				}
				redraw = true
				clear = true
			}
		case <-app.stop:
			return nil
		}
		select {
		case <-app.stop:
			return nil
		default:
		}
		if redraw {
			if clear || app.alwaysClear {
				screen.Clear()
			}
			screen.HideCursor()
			app.root.Draw(screen)
			screen.Show()
		}
	}
}

func (app *Application) Stop() {
	select {
	case app.stop <- struct{}{}:
	default:
	}
	<-app.waitForStop
}

func (app *Application) ForceStop() {
	app.screen.Fini()
	select {
	case app.stop <- struct{}{}:
	default:
	}
}

type suspendUpdate struct {
	wait func()
}

type redrawUpdate struct{}

type setRootUpdate struct {
	newRoot Component
}

func (app *Application) Suspend(wait func()) {
	app.updates <- suspendUpdate{wait}
}

func (app *Application) Redraw() {
	app.updates <- redrawUpdate{}
}

func (app *Application) SetRoot(view Component) {
	app.screenLock.RLock()
	defer app.screenLock.RUnlock()
	if app.screen != nil {
		app.updates <- setRootUpdate{view}
	} else {
		app.root = view
		focusable, ok := app.root.(Focusable)
		if ok {
			focusable.Focus()
		}
	}
}

// Screen returns the main tcell screen currently used in the app.
func (app *Application) Screen() tcell.Screen {
	app.screenLock.RLock()
	screen := app.screen
	app.screenLock.RUnlock()
	return screen
}

func (app *Application) SetAlwaysClear(always bool) {
	app.alwaysClear = always
}
