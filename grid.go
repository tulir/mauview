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

type gridChild struct {
	screen    *ProxyScreen
	relWidth  int
	relHeight int
	relX      int
	relY      int
	target    Component
}

func (child gridChild) Within(x, y int) bool {
	screen := child.screen
	return x >= screen.offsetX && x <= screen.offsetX+screen.width &&
		y >= screen.offsetY && y <= screen.offsetY+screen.height
}

func (child gridChild) Focus() {
	focusable, ok := child.target.(Focusable)
	if ok {
		focusable.Focus()
	}
}

func (child gridChild) Blur() {
	focusable, ok := child.target.(Focusable)
	if ok {
		focusable.Blur()
	}
}

type Grid struct {
	screen   Screen
	children []gridChild
	focused  *gridChild

	relWidth, relHeight         int
	prevAbsWidth, prevAbsHeight int

	columnWidths []int
	rowHeights   []int
}

func NewGrid(width, height int) *Grid {
	return &Grid{
		children:      []gridChild{},
		focused:       nil,
		relWidth:      width,
		relHeight:     height,
		prevAbsWidth:  -1,
		prevAbsHeight: -1,
		columnWidths:  make([]int, width),
		rowHeights:    make([]int, height),
	}
}

func (grid *Grid) AddComponent(comp Component, x, y, width, height int) *Grid {
	grid.children = append(grid.children, gridChild{
		screen:    &ProxyScreen{parent: grid.screen, style: tcell.StyleDefault},
		relWidth:  width,
		relHeight: height,
		relX:      x,
		relY:      y,
		target:    comp,
	})
	return grid
}

func (grid *Grid) RemoveComponent(comp Component) *Grid {
	for index := len(grid.children) - 1; index >= 0; index-- {
		if grid.children[index].target == comp {
			grid.children = append(grid.children[:index], grid.children[index+1:]...)
		}
	}
	return grid
}

func (grid *Grid) SetColumnWidth(col, width int) {
	grid.columnWidths[col] = width
}

func (grid *Grid) SetRowHeight(row, height int) {
	grid.rowHeights[row] = height
}

func sum(arr []int) (int, int) {
	sum := 0
	n := 0
	for _, i := range arr {
		sum += i
		if i <= 0 {
			n++
		}
	}
	return sum, n
}

func fillDynamic(arr []int, size, dynamicItems int) []int {
	if dynamicItems == 0 {
		return arr
	}
	part := size / dynamicItems
	remainder := size % dynamicItems
	newArr := make([]int, len(arr))
	for i, val := range arr {
		if val == 0 {
			if remainder > 0 {
				remainder--
				newArr[i] = part + 1
			} else {
				newArr[i] = part
			}
		} else {
			newArr[i] = val
		}
	}
	return newArr
}

func (grid *Grid) OnResize(width, height int) {
	absColWidth, dynamicColumns := sum(grid.columnWidths)
	columnWidths := fillDynamic(grid.columnWidths, width-absColWidth, dynamicColumns)
	absRowHeight, dynamicRows := sum(grid.rowHeights)
	rowHeights := fillDynamic(grid.rowHeights, height-absRowHeight, dynamicRows)
	for _, child := range grid.children {
		child.screen.offsetX, _ = sum(columnWidths[:child.relX])
		child.screen.offsetY, _ = sum(rowHeights[:child.relY])
		child.screen.width, _ = sum(columnWidths[child.relX : child.relX+child.relWidth])
		child.screen.height, _ = sum(rowHeights[child.relY : child.relY+child.relHeight])
	}
	grid.prevAbsWidth, grid.prevAbsHeight = width, height
}

func (grid *Grid) Draw(screen Screen) {
	width, height := screen.Size()
	if grid.prevAbsWidth != width || grid.prevAbsHeight != height {
		grid.OnResize(screen.Size())
	}
	screen.Clear()
	if screen != grid.screen {
		grid.screen = screen
		for _, child := range grid.children {
			child.screen.parent = screen
			child.target.Draw(child.screen)
		}
	} else {
		for _, child := range grid.children {
			child.target.Draw(child.screen)
		}
	}
}

func (grid *Grid) OnKeyEvent(event *tcell.EventKey) bool {
	if grid.focused != nil {
		return grid.focused.target.OnKeyEvent(event)
	}
	return false
}

func (grid *Grid) OnPasteEvent(event *tcell.EventPaste) bool {
	if grid.focused != nil {
		return grid.focused.target.OnPasteEvent(event)
	}
	return false
}

func (grid *Grid) OnMouseEvent(event *tcell.EventMouse) bool {
	if grid.focused != nil && grid.focused.Within(event.Position()) {
		return grid.focused.target.OnMouseEvent(event)
	}
	for _, child := range grid.children {
		if child.Within(event.Position()) {
			if event.Buttons() == tcell.Button1 {
				if grid.focused != nil {
					grid.focused.Blur()
				}
				grid.focused = &child
				grid.focused.Focus()
			}
			return child.target.OnMouseEvent(event)
		}
	}
	return false
}
