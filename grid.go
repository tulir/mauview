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

type gridChild struct {
	genericChild
	relWidth  int
	relHeight int
	relX      int
	relY      int
}

type Grid struct {
	screen   Screen
	children []*gridChild
	focused  *gridChild

	focusReceived bool

	prevWidth   int
	prevHeight  int
	forceResize bool

	columnWidths []int
	rowHeights   []int

	onFocusChanged func(from, to Component)
}

func NewGrid() *Grid {
	return &Grid{
		children:     []*gridChild{},
		focused:      nil,
		prevWidth:    -1,
		prevHeight:   -1,
		forceResize:  false,
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

func (form *Form) SetOnFocusChanged(fn func(from, to Component)) *Form {
	form.onFocusChanged = fn
	return form
}

func (grid *Grid) createChild(comp Component, x, y, width, height int) *gridChild {
	if x+width >= len(grid.columnWidths) {
		grid.columnWidths = extend(grid.columnWidths, x+width)
	}
	if y+height >= len(grid.rowHeights) {
		grid.rowHeights = extend(grid.rowHeights, y+height)
	}
	return &gridChild{
		genericChild: genericChild{
			screen: &ProxyScreen{Parent: grid.screen, Style: tcell.StyleDefault},
			target: comp,
		},
		relWidth:  width,
		relHeight: height,
		relX:      x,
		relY:      y,
	}
}

func (grid *Grid) addChild(child *gridChild) {
	if child.relX+child.relWidth >= len(grid.columnWidths) {
		grid.columnWidths = extend(grid.columnWidths, child.relX+child.relWidth)
	}
	if child.relY+child.relHeight >= len(grid.rowHeights) {
		grid.rowHeights = extend(grid.rowHeights, child.relY+child.relHeight)
	}
	grid.children = append(grid.children, child)
	grid.forceResize = true
}

func (grid *Grid) AddComponent(comp Component, x, y, width, height int) *Grid {
	grid.addChild(grid.createChild(comp, x, y, width, height))
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

func (grid *Grid) SetColumn(col, width int) *Grid {
	if col >= len(grid.columnWidths) {
		grid.columnWidths = extend(grid.columnWidths, col+1)
	}
	grid.columnWidths[col] = width
	return grid
}

func (grid *Grid) SetRow(row, height int) *Grid {
	if row >= len(grid.rowHeights) {
		grid.rowHeights = extend(grid.rowHeights, row+1)
	}
	grid.rowHeights[row] = height
	return grid
}

func (grid *Grid) SetColumns(columns []int) *Grid {
	grid.columnWidths = columns
	return grid
}

func (grid *Grid) SetRows(rows []int) *Grid {
	grid.rowHeights = rows
	return grid
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
		child.screen.OffsetX, _ = pnSum(columnWidths[:child.relX])
		child.screen.OffsetY, _ = pnSum(rowHeights[:child.relY])
		child.screen.Width, _ = pnSum(columnWidths[child.relX : child.relX+child.relWidth])
		child.screen.Height, _ = pnSum(rowHeights[child.relY : child.relY+child.relHeight])
	}
	grid.prevWidth, grid.prevHeight = width, height
}

func (grid *Grid) Draw(screen Screen) {
	width, height := screen.Size()
	if grid.forceResize || grid.prevWidth != width || grid.prevHeight != height {
		grid.OnResize(screen.Size())
	}
	grid.forceResize = false
	screenChanged := false
	if screen != grid.screen {
		grid.screen = screen
		screenChanged = true
	}
	for _, child := range grid.children {
		if screenChanged {
			child.screen.Parent = screen
		}
		if grid.focused == nil || child != grid.focused {
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

func (grid *Grid) setFocused(item *gridChild) {
	if grid.focused != nil {
		grid.focused.Blur()
	}
	var prevFocus, newFocus Component
	if grid.focused != nil {
		prevFocus = grid.focused.target
	}
	if item != nil {
		newFocus = item.target
	}
	grid.focused = item
	if grid.focusReceived && grid.focused != nil {
		grid.focused.Focus()
	}
	if grid.onFocusChanged != nil {
		grid.onFocusChanged(prevFocus, newFocus)
	}
}
func (grid *Grid) OnMouseEvent(event MouseEvent) bool {
	if grid.focused != nil && grid.focused.screen.IsInArea(event.Position()) {
		screen := grid.focused.screen
		return grid.focused.target.OnMouseEvent(OffsetMouseEvent(event, -screen.OffsetX, -screen.OffsetY))
	}
	for _, child := range grid.children {
		if child.screen.IsInArea(event.Position()) {
			focusChanged := false
			if event.Buttons() == tcell.Button1 && !event.HasMotion() {
				grid.setFocused(child)
				focusChanged = true
			}
			return child.target.OnMouseEvent(OffsetMouseEvent(event, -child.screen.OffsetX, -child.screen.OffsetY)) ||
				focusChanged

		}
	}
	if event.Buttons() == tcell.Button1 && !event.HasMotion() && grid.focused != nil {
		grid.setFocused(nil)
		return true
	}
	return false
}

func (grid *Grid) Focus() {
	grid.focusReceived = true
	if grid.focused != nil {
		grid.focused.Focus()
	}
}

func (grid *Grid) Blur() {
	if grid.focused != nil {
		grid.setFocused(nil)
	}
	grid.focusReceived = false
}
