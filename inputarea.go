// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Based on https://github.com/rivo/tview/blob/master/inputfield.go

package mauview

import (
	"strings"

	"github.com/mattn/go-runewidth"

	"maunium.net/go/tcell"
)

// InputArea is a multi-line user-editable text area.
type InputArea struct {
	// Cursor position as the runewidth from the start of the input area text.
	cursorOffsetW int
	// Cursor offset from the left of the input area.
	cursorOffsetX int
	// Cursor offset from the top of the text.
	cursorOffsetY int
	// Number of lines to offset rendering.
	viewOffsetY int

	// The start of the selection as the runewidth from the start of the input area text.
	selectionStartW int
	// The end of the selection.
	selectionEndW int

	// The text that was entered.
	text string

	// The text split into lines. Updated each during each render.
	lines []string

	// The text to be displayed in the input area when it is empty.
	placeholder string

	// The background color of the input area.
	fieldBackgroundColor tcell.Color
	// The text color of the input area.
	fieldTextColor tcell.Color
	// The text color of the placeholder.
	placeholderTextColor tcell.Color
	// The text color of selected text.
	selectionTextColor tcell.Color
	// The background color of selected text.
	selectionBackgroundColor tcell.Color

	// Whether or not to enable vim-style keybindings.
	vimBindings bool

	// Whether or not the input area is focused.
	focused bool

	// An optional function which is called when the input has changed.
	changed func(text string)
}

// NewInputArea returns a new input field.
func NewInputArea() *InputArea {
	return &InputArea{
		fieldBackgroundColor:     Styles.PrimitiveBackgroundColor,
		fieldTextColor:           Styles.PrimaryTextColor,
		placeholderTextColor:     Styles.SecondaryTextColor,
		selectionTextColor:       Styles.PrimaryTextColor,
		selectionBackgroundColor: Styles.ContrastBackgroundColor,

		vimBindings: false,
		focused:     false,

		selectionEndW:   -1,
		selectionStartW: -1,
	}
}

// SetText sets the current text of the input field.
func (field *InputArea) SetText(text string) *InputArea {
	field.text = text
	if field.changed != nil {
		field.changed(text)
	}
	return field
}

// SetTextAndMoveCursor sets the current text of the input field and moves the cursor with the width difference.
func (field *InputArea) SetTextAndMoveCursor(text string) *InputArea {
	oldWidth := iaStringWidth(field.text)
	field.text = text
	newWidth := iaStringWidth(field.text)
	if oldWidth != newWidth {
		field.cursorOffsetW += newWidth - oldWidth
	}
	if field.changed != nil {
		field.changed(field.text)
	}
	return field
}

// GetText returns the current text of the input field.
func (field *InputArea) GetText() string {
	return field.text
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (field *InputArea) SetPlaceholder(text string) *InputArea {
	field.placeholder = text
	return field
}

// SetFieldBackgroundColor sets the background color of the input area.
func (field *InputArea) SetFieldBackgroundColor(color tcell.Color) *InputArea {
	field.fieldBackgroundColor = color
	return field
}

// SetFieldTextColor sets the text color of the input area.
func (field *InputArea) SetFieldTextColor(color tcell.Color) *InputArea {
	field.fieldTextColor = color
	return field
}

// SetPlaceholderExtColor sets the text color of placeholder text.
func (field *InputArea) SetPlaceholderExtColor(color tcell.Color) *InputArea {
	field.placeholderTextColor = color
	return field
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (field *InputArea) SetChangedFunc(handler func(text string)) *InputArea {
	field.changed = handler
	return field
}

// GetTextHeight returns the number of lines in the text during the previous render.
func (field *InputArea) GetTextHeight() int {
	return len(field.lines)
}

func matchBoundaryPattern(extract string) string {
	matches := boundaryPattern.FindAllStringIndex(extract, -1)
	if len(matches) > 0 {
		if match := matches[len(matches)-1]; len(match) >= 2 {
			if until := match[1]; until < len(extract) {
				extract = extract[:until]
			}
		}
	}
	return extract
}

func (field *InputArea) recalculateCursorOffset() {
	cursorOffsetW := 0
	for i, str := range field.lines {
		ln := iaStringWidth(str)
		if i < field.cursorOffsetY {
			cursorOffsetW += ln
		} else {
			if ln == 0 {
				break
			} else if str[len(str)-1] == '\n' {
				ln--
			}
			if field.cursorOffsetX < ln {
				cursorOffsetW += field.cursorOffsetX
			} else {
				cursorOffsetW += ln
			}
			break
		}
	}
	field.cursorOffsetW = cursorOffsetW
	textWidth := iaStringWidth(field.text)
	if field.cursorOffsetW > textWidth {
		field.cursorOffsetW = textWidth
		field.recalculateCursorPos()
	}
}

func (field *InputArea) recalculateCursorPos() {
	cursorOffsetY := 0
	cursorOffsetX := field.cursorOffsetW
	for i, str := range field.lines {
		if cursorOffsetX >= iaStringWidth(str) {
			cursorOffsetX -= iaStringWidth(str)
		} else {
			cursorOffsetY = i
			break
		}
	}
	field.cursorOffsetX = cursorOffsetX
	field.cursorOffsetY = cursorOffsetY
}

func (field *InputArea) prepareText(width int) {
	var lines []string
	if len(field.text) == 0 {
		field.lines = lines
		return
	}
	forcedLinebreaks := strings.Split(field.text, "\n")
	for _, str := range forcedLinebreaks {
		str = str + "\n"
		// Adapted from tview/textview.go#reindexBuffer()
		for len(str) > 0 {
			extract := iaSubstringBefore(str, width-1)
			if len(extract) < len(str) {
				if spaces := spacePattern.FindStringIndex(str[len(extract):]); spaces != nil && spaces[0] == 0 {
					extract = str[:len(extract)+spaces[1]]
				}
				extract = matchBoundaryPattern(extract)
			}
			lines = append(lines, extract)
			str = str[len(extract):]
		}
	}
	field.lines = lines
}

func (field *InputArea) updateViewOffset(height int) {
	if field.viewOffsetY < 0 {
		field.viewOffsetY = 0
	} else if len(field.lines) > height && field.viewOffsetY+height > len(field.lines) {
		field.viewOffsetY = len(field.lines) - height
	}
	if field.cursorOffsetY-field.viewOffsetY < 0 {
		field.viewOffsetY = field.cursorOffsetY
	} else if field.cursorOffsetY >= field.viewOffsetY+height {
		field.viewOffsetY = field.cursorOffsetY - height + 1
	}
}

// drawText draws the text and the cursor.
func (field *InputArea) drawText(screen Screen) {
	width, height := screen.Size()
	if len(field.lines) == 0 {
		if len(field.placeholder) > 0 {
			Print(screen, field.placeholder, 0, 0, width, AlignLeft, field.placeholderTextColor)
		}
		return
	}
	rwOffset := 0
	for y := 0; y <= field.viewOffsetY+height && y < len(field.lines); y++ {
		if y < field.viewOffsetY {
			rwOffset += iaStringWidth(field.lines[y])
			continue
		}
		x := 0
		for _, ch := range []rune(field.lines[y]) {
			w := iaRuneWidth(ch)
			_, _, style, _ := screen.GetContent(x, y)
			style = style.Foreground(field.fieldTextColor)
			if rwOffset >= field.selectionStartW && rwOffset < field.selectionEndW {
				style = style.Foreground(field.selectionTextColor).Background(field.selectionBackgroundColor)
			}
			rwOffset += w
			for w > 0 {
				screen.SetContent(x, y-field.viewOffsetY, ch, nil, style)
				x++
				w--
			}
		}
	}
}

// Draw draws this primitive onto the screen.
func (field *InputArea) Draw(screen Screen) {
	width, height := screen.Size()
	if height < 1 || width < 1 {
		return
	}

	screen.SetStyle(tcell.StyleDefault.Background(field.fieldBackgroundColor))
	screen.Clear()
	field.prepareText(width)
	field.recalculateCursorPos()
	field.updateViewOffset(height)
	field.drawText(screen)
	if field.focused && field.selectionEndW == -1 {
		screen.ShowCursor(field.cursorOffsetX, field.cursorOffsetY-field.viewOffsetY)
	}
}

func iaRuneWidth(ch rune) int {
	if ch == '\n' {
		return 1
	}
	return runewidth.RuneWidth(ch)
}

func iaStringWidth(s string) (width int) {
	w := runewidth.StringWidth(s)
	for _, ch := range s {
		if ch == '\n' {
			w++
		}
	}
	return w
}

func iaSubstringBefore(s string, w int) string {
	if iaStringWidth(s) <= w {
		return s
	}
	r := []rune(s)
	//tw := iaStringWidth(tail)
	//w -= tw
	width := 0
	i := 0
	for ; i < len(r); i++ {
		cw := iaRuneWidth(r[i])
		if width+cw > w {
			break
		}
		width += cw
	}
	return string(r[0:i]) // + tail
}

func (field *InputArea) TypeRune(ch rune) {
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	right := field.text[len(left):]
	field.text = left + string(ch) + right
	field.cursorOffsetW += iaRuneWidth(ch)
}

func (field *InputArea) MoveCursorLeft(moveWord, extendSelection bool) {
	before := iaSubstringBefore(field.text, field.cursorOffsetW)
	var diff int
	if moveWord {
		diff = -iaStringWidth(lastWord.FindString(before))
	} else if len(before) > 0 {
		beforeRunes := []rune(before)
		char := beforeRunes[len(beforeRunes)-1]
		diff = -iaRuneWidth(char)
	}
	if extendSelection {
		field.extendSelection(diff)
	} else {
		field.moveCursor(diff)
	}
}

func (field *InputArea) MoveCursorRight(moveWord, extendSelection bool) {
	before := iaSubstringBefore(field.text, field.cursorOffsetW)
	after := field.text[len(before):]
	var diff int
	if moveWord {
		diff = +iaStringWidth(firstWord.FindString(after))
	} else if len(after) > 0 {
		char := []rune(after)[0]
		diff = +iaRuneWidth(char)
	}
	if extendSelection {
		field.extendSelection(diff)
	} else {
		field.moveCursor(diff)
	}
}

func (field *InputArea) moveCursor(diff int) {
	field.selectionEndW = -1
	field.selectionStartW = -1
	field.cursorOffsetW += diff
}

func (field *InputArea) extendSelection(diff int) {
	if field.selectionEndW == -1 {
		field.selectionStartW = field.cursorOffsetW
		field.selectionEndW = field.selectionStartW + diff
	} else if field.cursorOffsetW == field.selectionEndW {
		field.selectionEndW += diff
	} else if field.cursorOffsetW == field.selectionStartW {
		field.selectionStartW += diff
	}
	field.cursorOffsetW += diff
	if field.selectionStartW > field.selectionEndW {
		field.selectionStartW, field.selectionEndW = field.selectionEndW, field.selectionStartW
	}
}

func (field *InputArea) MoveCursorUp(extendSelection bool) {
	pX, pY := field.cursorOffsetX, field.cursorOffsetY
	if field.cursorOffsetY > 0 {
		field.cursorOffsetY--
		lineWidth := iaStringWidth(field.lines[field.cursorOffsetY])
		if lineWidth < field.cursorOffsetX {
			field.cursorOffsetX = lineWidth
		}
	}
	if extendSelection {
		prevLineBefore := iaSubstringBefore(field.lines[pY], pX)
		curLineBefore := iaSubstringBefore(field.lines[field.cursorOffsetY], field.cursorOffsetX)
		curLineAfter := field.lines[field.cursorOffsetY][len(curLineBefore):]
		field.extendSelection(-iaStringWidth(curLineAfter + prevLineBefore))
	} else {
		field.selectionStartW = -1
		field.selectionEndW = -1
	}
	field.recalculateCursorOffset()
}

func (field *InputArea) MoveCursorDown(extendSelection bool) {
	pX, pY := field.cursorOffsetX, field.cursorOffsetY
	if field.cursorOffsetY < len(field.lines)-1 {
		field.cursorOffsetY++
		lineWidth := iaStringWidth(field.lines[field.cursorOffsetY])
		if lineWidth < field.cursorOffsetX {
			field.cursorOffsetX = lineWidth
		}
	} else if field.cursorOffsetY == len(field.lines)-1 {
		lineWidth := iaStringWidth(field.lines[field.cursorOffsetY])
		field.cursorOffsetX = lineWidth
	}
	if extendSelection {
		prevLineBefore := iaSubstringBefore(field.lines[pY], pX)
		prevLineAfter := field.lines[pY][len(prevLineBefore):]
		curLineBefore := iaSubstringBefore(field.lines[field.cursorOffsetY], field.cursorOffsetX)
		field.extendSelection(iaStringWidth(prevLineAfter + curLineBefore))
	} else {
		field.selectionStartW = -1
		field.selectionEndW = -1
	}
	field.recalculateCursorOffset()
}

func (field *InputArea) MoveCursorPos(x, y int, moveSelection bool) {
	field.cursorOffsetX = x
	field.cursorOffsetY = y
	if field.cursorOffsetY > len(field.lines) {
		field.cursorOffsetY = len(field.lines) - 1
	}
	prevOffset := field.cursorOffsetW
	field.recalculateCursorOffset()
	if moveSelection {
		if field.selectionEndW == -1 {
			field.selectionStartW = prevOffset
			field.selectionEndW = field.cursorOffsetW
		} else if prevOffset == field.selectionEndW {
			field.selectionEndW = field.cursorOffsetW
		} else {
			field.selectionStartW = field.cursorOffsetW
		}
		if field.selectionStartW > field.selectionEndW {
			field.selectionStartW, field.selectionEndW = field.selectionEndW, field.selectionStartW
		}
	} else {
		field.selectionStartW = -1
		field.selectionEndW = -1
	}
}

func (field *InputArea) RemoveNextCharacter() {
	if field.selectionEndW > 0 {
		field.RemoveSelection()
		return
	} else if field.cursorOffsetW >= iaStringWidth(field.text) {
		return
	}
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	// Take everything after the left part minus the first character.
	right := string([]rune(field.text[len(left):])[1:])

	field.text = left + right
}

func (field *InputArea) RemovePreviousWord() {
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	replacement := lastWord.ReplaceAllString(left, "")
	field.text = replacement + field.text[len(left):]
	field.cursorOffsetW = iaStringWidth(replacement)
}

func (field *InputArea) RemoveSelection() {
	leftLeft := iaSubstringBefore(field.text, field.selectionStartW)
	rightLeft := iaSubstringBefore(field.text, field.selectionEndW)
	rightRight := field.text[len(rightLeft):]
	field.text = leftLeft + rightRight
	if field.cursorOffsetW == field.selectionEndW {
		field.cursorOffsetW -= iaStringWidth(rightLeft[len(leftLeft):])
	}
	field.selectionEndW = -1
	field.selectionStartW = -1
}

func (field *InputArea) RemovePreviousCharacter() {
	if field.selectionEndW > 0 {
		field.RemoveSelection()
		return
	} else if field.cursorOffsetW == 0 {
		return
	}
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	right := field.text[len(left):]

	// Take everything before the right part minus the last character.
	leftRunes := []rune(left)
	leftRunes = leftRunes[0 : len(leftRunes)-1]
	left = string(leftRunes)

	// Figure out what character was removed to correctly decrease cursorOffset.
	removedChar := field.text[len(left) : len(field.text)-len(right)]
	field.text = left + right
	field.cursorOffsetW -= iaStringWidth(removedChar)
}

func (field *InputArea) Clear() {
	field.text = ""
	field.cursorOffsetW = 0
	field.cursorOffsetX = 0
	field.cursorOffsetY = 0
	field.selectionEndW = -1
	field.selectionStartW = -1
	field.viewOffsetY = 0
}

func (field *InputArea) SelectAll() {
	field.selectionStartW = 0
	field.selectionEndW = iaStringWidth(field.text)
	field.cursorOffsetW = field.selectionEndW
}

func (field *InputArea) handleInputChanges(originalText string) {
	// Trigger changed events.
	if field.text != originalText && field.changed != nil {
		field.changed(field.text)
	}

	// Make sure cursor offset is valid
	if field.cursorOffsetW < 0 {
		field.cursorOffsetW = 0
	}
	textWidth := iaStringWidth(field.text)
	if field.cursorOffsetW > textWidth {
		field.cursorOffsetW = textWidth
	}
	if field.selectionEndW > textWidth {
		field.selectionEndW = textWidth
	}
	if field.selectionEndW <= field.selectionStartW {
		field.selectionStartW = -1
		field.selectionEndW = -1
	}
}

func (field *InputArea) OnPasteEvent(event PasteEvent) bool {
	defer field.handleInputChanges(field.text)
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	right := field.text[len(left):]
	field.text = left + event.Text() + right
	field.cursorOffsetW += iaStringWidth(event.Text())
	return true
}

func (field *InputArea) OnKeyEvent(event KeyEvent) bool {
	defer field.handleInputChanges(field.text)

	hasMod := func(mod tcell.ModMask) bool {
		return event.Modifiers()&mod != 0
	}

	// Process key event.
	switch key := event.Key(); key {
	case tcell.KeyRune:
		field.TypeRune(event.Rune())
	case tcell.KeyEnter:
		field.TypeRune('\n')
	case tcell.KeyLeft, tcell.KeyCtrlLeft, tcell.KeyShiftLeft, tcell.KeyCtrlShiftLeft:
		field.MoveCursorLeft(hasMod(tcell.ModCtrl), hasMod(tcell.ModShift))
	case tcell.KeyRight, tcell.KeyCtrlRight, tcell.KeyShiftRight, tcell.KeyCtrlShiftRight:
		field.MoveCursorRight(hasMod(tcell.ModCtrl), hasMod(tcell.ModShift))
	case tcell.KeyUp, tcell.KeyShiftUp:
		field.MoveCursorUp(hasMod(tcell.ModShift))
	case tcell.KeyDown, tcell.KeyShiftDown:
		field.MoveCursorDown(hasMod(tcell.ModShift))
	case tcell.KeyDelete:
		field.RemoveNextCharacter()
	case tcell.KeyCtrlU:
		if field.vimBindings {
			field.Clear()
		}
	case tcell.KeyCtrlW:
		if field.vimBindings {
			field.RemovePreviousWord()
		}
	case tcell.KeyCtrlA:
		if !field.vimBindings {
			field.SelectAll()
		}
	case tcell.KeyBackspace:
		field.RemovePreviousWord()
	case tcell.KeyBackspace2:
		field.RemovePreviousCharacter()
	default:
		return false
	}
	return true
}

func (field *InputArea) Focus() {
	field.focused = true
}

func (field *InputArea) Blur() {
	field.focused = false
}

func (field *InputArea) OnMouseEvent(event MouseEvent) bool {
	switch event.Buttons() {
	case tcell.Button1:
		cursorX, cursorY := event.Position()
		field.MoveCursorPos(cursorX, field.viewOffsetY+cursorY, event.HasMotion())
	case tcell.WheelDown:
		field.MoveCursorDown(false)
	case tcell.WheelUp:
		field.MoveCursorUp(false)
	default:
		return false
	}
	return true
}
