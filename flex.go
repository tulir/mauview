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

type FlexDirection int

const (
	FlexRow FlexDirection = iota
	FlexColumn
)

type flexChild struct {
	genericChild
	size int
}

type Flex struct {
	direction FlexDirection
	children  []flexChild
	focused   *flexChild
}

func NewFlex() *Flex {
	return &Flex{
		children:  []flexChild{},
		focused:   nil,
		direction: FlexColumn,
	}
}

func (flex *Flex) SetDirection(direction FlexDirection) *Flex {
	flex.direction = direction
	return flex
}

func (flex *Flex) AddFixedComponent(comp Component, size int) *Flex {
	flex.AddProportionalComponent(comp, -size)
	return flex
}

func (flex *Flex) AddProportionalComponent(comp Component, size int) *Flex {
	flex.children = append(flex.children, flexChild{
		genericChild: genericChild{
			target: comp,
			screen: &ProxyScreen{Style: tcell.StyleDefault},
		},
		size: -size,
	})
	return flex
}

func (flex *Flex) RemoveComponent(comp Component) *Flex {
	for index := len(flex.children) - 1; index >= 0; index-- {
		if flex.children[index].target == comp {
			flex.children = append(flex.children[:index], flex.children[index+1:]...)
		}
	}
	return flex
}

func (flex *Flex) Draw(screen Screen) {
	width, height := screen.Size()
	relTotalSize := width
	if flex.direction == FlexRow {
		relTotalSize = height
	}
	relParts := 0
	for _, child := range flex.children {
		if child.size > 0 {
			relTotalSize -= child.size
		} else {
			relParts -= child.size
		}

	}
	offset := 0
	for _, child := range flex.children {
		child.screen.Parent = screen
		size := child.size
		if size < 0 {
			size = relTotalSize * (-size) / relParts
		}
		if flex.direction == FlexRow {
			child.screen.Height = size
			child.screen.Width = width
			child.screen.OffsetY = offset
			child.screen.OffsetX = 0
		} else {
			child.screen.Height = height
			child.screen.Width = size
			child.screen.OffsetY = 0
			child.screen.OffsetX = offset
		}
		offset += size
		if flex.focused == nil || child != *flex.focused {
			child.target.Draw(child.screen)
		}
	}
	if flex.focused != nil {
		flex.focused.target.Draw(flex.focused.screen)
	}
}

func (flex *Flex) OnKeyEvent(event KeyEvent) bool {
	if flex.focused != nil {
		return flex.focused.target.OnKeyEvent(event)
	}
	return false
}

func (flex *Flex) OnPasteEvent(event PasteEvent) bool {
	if flex.focused != nil {
		return flex.focused.target.OnPasteEvent(event)
	}
	return false
}

func (flex *Flex) SetFocused(comp Component) {
	for _, child := range flex.children {
		if child.target == comp {
			flex.focused = &child
			flex.focused.Focus()
		}
	}
}

func (flex *Flex) OnMouseEvent(event MouseEvent) bool {
	if flex.focused != nil && flex.focused.screen.IsInArea(event.Position()) {
		screen := flex.focused.screen
		return flex.focused.target.OnMouseEvent(OffsetMouseEvent(event, -screen.OffsetX, -screen.OffsetY))
	}
	for _, child := range flex.children {
		if child.screen.IsInArea(event.Position()) {
			focusChanged := false
			if event.Buttons() == tcell.Button1 && !event.HasMotion() {
				if flex.focused != nil {
					flex.focused.Blur()
				}
				flex.focused = &child
				flex.focused.Focus()
				focusChanged = true
			}
			return child.target.OnMouseEvent(OffsetMouseEvent(event, -child.screen.OffsetX, -child.screen.OffsetY)) ||
				focusChanged

		}
	}
	if event.Buttons() == tcell.Button1 && flex.focused != nil && !event.HasMotion() {
		flex.focused.Blur()
		flex.focused = nil
		return true
	}
	return false
}

func (flex *Flex) Focus() {}

func (flex *Flex) Blur() {
	if flex.focused != nil {
		flex.focused.Blur()
		flex.focused = nil
	}
}
