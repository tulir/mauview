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
	genericChild
	relWidth  int
	relHeight int
	relX      int
	relY      int
}

type Grid struct {
	screen   Screen
	children []gridChild
	focused  *gridChild

	prevWidth  int
	prevHeight int

	columnWidths []int
	rowHeights   []int
}

func NewGrid() *Grid {
	return &Grid{
		children:     []gridChild{},
		focused:      nil,
		prevWidth:    -1,
		prevHeight:   -1,
		columnWidths: []int{-1},
		rowHeights:   []int{-1},
	}
}

func extend(arr []int, newSize int) []int {
	newArr := make([]int, newSize)
	copy(newArr, arr)
	for i := len(arr); i < len(newArr); i++ {
		newArr[i] = -1
	}
	return newArr
}

func (grid *Grid) AddComponent(comp Component, x, y, width, height int) *Grid {
	if x+width >= len(grid.columnWidths) {
		grid.columnWidths = extend(grid.columnWidths, x+width)
	}
	if y+height >= len(grid.rowHeights) {
		grid.rowHeights = extend(grid.rowHeights, y+height)
	}
	grid.children = append(grid.children, gridChild{
		genericChild: genericChild{
			screen: &ProxyScreen{parent: grid.screen, style: tcell.StyleDefault},
			target: comp,
		},
		relWidth:  width,
		relHeight: height,
		relX:      x,
		relY:      y,
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

func (grid *Grid) SetColumn(col, width int) {
	if col >= len(grid.columnWidths) {
		grid.columnWidths = extend(grid.columnWidths, col+1)
	}
	grid.columnWidths[col] = width
}

func (grid *Grid) SetRow(row, height int) {
	if row >= len(grid.rowHeights) {
		grid.rowHeights = extend(grid.rowHeights, row+1)
	}
	grid.rowHeights[row] = height
}

func (grid *Grid) SetColumns(columns []int) {
	grid.columnWidths = columns
}

func (grid *Grid) SetRows(rows []int) {
	grid.rowHeights = rows
}

func pnSum(arr []int) (int, int) {
	positive := 0
	negative := 0
	for _, i := range arr {
		if i < 0 {
			negative -= i
		} else {
			positive += i
		}
	}
	return positive, negative
}

func fillDynamic(arr []int, size, dynamicItems int) []int {
	if dynamicItems == 0 {
		return arr
	}
	part := size / dynamicItems
	remainder := size % dynamicItems
	newArr := make([]int, len(arr))
	for i, val := range arr {
		if val < 0 {
			newArr[i] = part * -val
			if remainder > 0 {
				remainder--
				newArr[i]++
			}
		} else {
			newArr[i] = val
		}
	}
	return newArr
}

func (grid *Grid) OnResize(width, height int) {
	absColWidth, dynamicColumns := pnSum(grid.columnWidths)
	columnWidths := fillDynamic(grid.columnWidths, width-absColWidth, dynamicColumns)
	absRowHeight, dynamicRows := pnSum(grid.rowHeights)
	rowHeights := fillDynamic(grid.rowHeights, height-absRowHeight, dynamicRows)
	for _, child := range grid.children {
		child.screen.offsetX, _ = pnSum(columnWidths[:child.relX])
		child.screen.offsetY, _ = pnSum(rowHeights[:child.relY])
		child.screen.width, _ = pnSum(columnWidths[child.relX : child.relX+child.relWidth])
		child.screen.height, _ = pnSum(rowHeights[child.relY : child.relY+child.relHeight])
	}
	grid.prevWidth, grid.prevHeight = width, height
}

func (grid *Grid) Draw(screen Screen) {
	width, height := screen.Size()
	if grid.prevWidth != width || grid.prevHeight != height {
		grid.OnResize(screen.Size())
	}
	screen.Clear()
	screenChanged := false
	if screen != grid.screen {
		grid.screen = screen
		screenChanged = true
	}
	for _, child := range grid.children {
		if screenChanged {
			child.screen.parent = screen
		}
		if grid.focused == nil || child != *grid.focused {
			child.target.Draw(child.screen)
		}
	}
	if grid.focused != nil {
		grid.focused.target.Draw(grid.focused.screen)
	}
}

func (grid *Grid) OnKeyEvent(event KeyEvent) bool {
	if grid.focused != nil {
		return grid.focused.target.OnKeyEvent(event)
	}
	return false
}

func (grid *Grid) OnPasteEvent(event PasteEvent) bool {
	if grid.focused != nil {
		return grid.focused.target.OnPasteEvent(event)
	}
	return false
}

func (grid *Grid) OnMouseEvent(event MouseEvent) bool {
	if grid.focused != nil && grid.focused.Within(event.Position()) {
		screen := grid.focused.screen
		return grid.focused.target.OnMouseEvent(OffsetMouseEvent(event, -screen.offsetX, -screen.offsetY))
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
			return child.target.OnMouseEvent(OffsetMouseEvent(event, -child.screen.offsetX, -child.screen.offsetY)) ||
				event.Buttons() == tcell.Button1

		}
	}
	if event.Buttons() == tcell.Button1 && grid.focused != nil {
		grid.focused.Blur()
		grid.focused = nil
		return true
	}
	return false
}

func (grid *Grid) Focus() {}

func (grid *Grid) Blur() {
	if grid.focused != nil {
		grid.focused.Blur()
		grid.focused = nil
	}
}
