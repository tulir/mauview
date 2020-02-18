// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/zyedidia/clipboard"

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
	// Number of lines (from top) to offset rendering.
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
	// Whether or not text should be automatically copied to the primary clipboard when selected.
	// Most apps on Linux work this way.
	copySelection bool

	// Whether or not the input area is focused.
	focused bool

	drawPrepared bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user presses tab.
	tabComplete func(text string, pos int)
	// An optional function which is called when the user presses the down arrow at the end of the input area.
	pressKeyDownAtEnd func()
	// An optional function which is called when the user presses the up arrow at the beginning of the input area.
	pressKeyUpAtStart func()

	// Change history for undo/redo functionality.
	history []*inputAreaSnapshot
	// Current position in the history array for redo functionality.
	historyPtr int
	// Maximum number of history snapshots to keep.
	historyMaxSize int
	// Maximum delay (ms) between changes to edit the previous snapshot instead of creating a new one.
	historyMaxEditDelay int64
	// Maximum age (ms) of the previous snapshot to edit the previous snapshot instead of craeting a new one.
	historyMaxSnapshotAge int64

	// Timestamp of the last click used for detecting double clicks.
	lastClick int64
	// Position of the last click used for detecting double clicks.
	lastClickX int
	lastClickY int
	// Number of clicks done within doubleClickTimeout of eachother.
	clickStreak int
	// Maximum delay (ms) between clicks to count as a double click.
	doubleClickTimeout int64

	// The previous word start and end X position that the mouse was dragged over when selecting words at a time.
	// Used to detect if the mouse is still over the same word.
	lastWordSelectionExtendXStart int
	lastWordSelectionExtendXEnd   int
	// The position where the current selection streak started.
	// Used to properly handle the user selecting text backwards.
	selectionStreakStartWStart int
	selectionStreakStartWEnd   int
	selectionStreakStartXStart int
	selectionStreakStartY      int
}

// NewInputArea returns a new input field.
func NewInputArea() *InputArea {
	return &InputArea{
		fieldBackgroundColor:     Styles.PrimitiveBackgroundColor,
		fieldTextColor:           Styles.PrimaryTextColor,
		placeholderTextColor:     Styles.SecondaryTextColor,
		selectionTextColor:       Styles.PrimaryTextColor,
		selectionBackgroundColor: Styles.ContrastBackgroundColor,

		vimBindings:   false,
		copySelection: true,
		focused:       false,

		selectionEndW:   -1,
		selectionStartW: -1,

		history:    []*inputAreaSnapshot{{"", 0, 0, 0, true}},
		historyPtr: 0,

		historyMaxSize:        256,
		historyMaxEditDelay:   1 * 1000,
		historyMaxSnapshotAge: 3 * 1000,

		lastClick:          0,
		doubleClickTimeout: 1 * 500,
	}
}

// SetText sets the current text of the input field.
func (field *InputArea) SetText(text string) *InputArea {
	field.text = text
	if field.changed != nil {
		field.changed(text)
	}
	field.snapshot()
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
	field.snapshot()
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

// SetBackgroundColor sets the background color of the input area.
func (field *InputArea) SetBackgroundColor(color tcell.Color) *InputArea {
	field.fieldBackgroundColor = color
	return field
}

// SetTextColor sets the text color of the input area.
func (field *InputArea) SetTextColor(color tcell.Color) *InputArea {
	field.fieldTextColor = color
	return field
}

// SetPlaceholderTextColor sets the text color of placeholder text.
func (field *InputArea) SetPlaceholderTextColor(color tcell.Color) *InputArea {
	field.placeholderTextColor = color
	return field
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (field *InputArea) SetChangedFunc(handler func(text string)) *InputArea {
	field.changed = handler
	return field
}

func (field *InputArea) SetTabCompleteFunc(handler func(text string, cursorOffset int)) *InputArea {
	field.tabComplete = handler
	return field
}

func (field *InputArea) SetPressKeyUpAtStartFunc(handler func()) *InputArea {
	field.pressKeyUpAtStart = handler
	return field
}

func (field *InputArea) SetPressKeyDownAtEndFunc(handler func()) *InputArea {
	field.pressKeyDownAtEnd = handler
	return field
}

// GetTextHeight returns the number of lines in the text during the previous render.
func (field *InputArea) GetTextHeight() int {
	return len(field.lines)
}

// inputAreaSnapshot is a single history snapshot of the input area state.
type inputAreaSnapshot struct {
	text          string
	cursorOffsetW int
	origTimestamp int64
	editTimestamp int64
	locked        bool
}

func millis() int64 {
	return time.Now().UnixNano() / 1e6
}

// Snapshot saves the current editor state into undo history.
func (field *InputArea) snapshot() {
	cur := field.history[field.historyPtr]
	now := millis()
	if cur.locked || now > cur.editTimestamp+field.historyMaxEditDelay || now > cur.origTimestamp+field.historyMaxSnapshotAge {
		newSnapshot := &inputAreaSnapshot{
			text:          field.text,
			cursorOffsetW: field.cursorOffsetW,
			origTimestamp: now,
			editTimestamp: now,
		}
		if len(field.history) >= field.historyMaxSize {
			field.history = append(field.history[1:field.historyPtr+1], newSnapshot)
		} else {
			field.history = append(field.history[0:field.historyPtr+1], newSnapshot)
			field.historyPtr++
		}
	} else {
		cur.text = field.text
		cur.cursorOffsetW = field.cursorOffsetW
		cur.editTimestamp = now
	}
}

// Redo reverses an undo.
func (field *InputArea) Redo() {
	if field.historyPtr >= len(field.history)-1 {
		return
	}
	field.historyPtr++
	newCur := field.history[field.historyPtr]
	newCur.locked = true
	field.text = newCur.text
	field.cursorOffsetW = newCur.cursorOffsetW
}

// Undo reverses the input area to the previous history snapshot.
func (field *InputArea) Undo() {
	if field.historyPtr == 0 {
		return
	}
	field.historyPtr--
	newCur := field.history[field.historyPtr]
	newCur.locked = true
	field.text = newCur.text
	field.cursorOffsetW = newCur.cursorOffsetW
}

// recalculateCursorOffset recalculates the runewidth cursor offset based on the X and Y cursor offsets.
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

// recalculateCursorPos recalculates the X and Y cursor offsets based on the runewidth cursor offset.
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

// prepareText splits the text into lines that fit the input area.
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

// updateViewOffset updates the view offset so that:
//   * it is not negative
//   * it is not unnecessarily high
//   * the cursor is within the rendered area
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
	defaultStyle := tcell.StyleDefault.Foreground(field.fieldTextColor).Background(field.fieldBackgroundColor)
	highlightStyle := defaultStyle.Foreground(field.selectionTextColor).Background(field.selectionBackgroundColor)
	rwOffset := 0
	for y := 0; y <= field.viewOffsetY+height && y < len(field.lines); y++ {
		if y < field.viewOffsetY {
			rwOffset += iaStringWidth(field.lines[y])
			continue
		}
		x := 0
		for _, ch := range []rune(field.lines[y]) {
			w := iaRuneWidth(ch)
			var style tcell.Style
			if rwOffset >= field.selectionStartW && rwOffset < field.selectionEndW {
				style = highlightStyle
			} else {
				style = defaultStyle
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

func (field *InputArea) PrepareDraw(width int) {
	field.prepareText(width)
	field.recalculateCursorPos()
}

// Draw draws this primitive onto the screen.
func (field *InputArea) Draw(screen Screen) {
	width, height := screen.Size()
	if height < 1 || width < 1 {
		return
	}

	if !field.drawPrepared {
		field.PrepareDraw(width)
	}
	field.updateViewOffset(height)
	screen.SetStyle(tcell.StyleDefault.Background(field.fieldBackgroundColor))
	screen.Clear()
	field.drawText(screen)
	if field.focused && field.selectionEndW == -1 {
		screen.ShowCursor(field.cursorOffsetX, field.cursorOffsetY-field.viewOffsetY)
	}
	field.drawPrepared = false
}

func iaRuneWidth(ch rune) int {
	if ch == '\n' {
		return 1
	}
	return runewidth.RuneWidth(ch)
}

func iaStringWidth(s string) (width int) {
	w := StringWidth(s)
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

// TypeRune inserts the given rune at the current cursor position.
func (field *InputArea) TypeRune(ch rune) {
	var left, right string
	if field.selectionEndW != -1 {
		left = iaSubstringBefore(field.text, field.selectionStartW)
		rightLeft := iaSubstringBefore(field.text, field.selectionEndW)
		right = field.text[len(rightLeft):]
		field.cursorOffsetW = field.selectionStartW
	} else {
		left = iaSubstringBefore(field.text, field.cursorOffsetW)
		right = field.text[len(left):]
	}
	field.text = left + string(ch) + right
	field.cursorOffsetW += iaRuneWidth(ch)
	field.selectionEndW = -1
	field.selectionStartW = -1
}

// MoveCursorLeft moves the cursor left.
//
// If moveWord is true, the cursor moves a whole word to the left.
//
// If extendSelection is true, the selection is either extended to the left if the cursor is on the left side of the
// selection or retracted from the right if the cursor is on the right side. If there is no existing selection, the
// selection will be created towards the left of the cursor.
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

// MoveCursorLeft moves the cursor right.
//
// If moveWord is true, the cursor moves a whole word to the right.
//
// If extendSelection is true, the selection is either extended to the right if the cursor is on the right side of the
// selection or retracted from the left if the cursor is on the left side. If there is no existing selection, the
// selection will be created towards the right of the cursor.
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

// moveCursor resets the selection and adjusts the runewidth cursor offset.
func (field *InputArea) moveCursor(diff int) {
	field.selectionEndW = -1
	field.selectionStartW = -1
	field.cursorOffsetW += diff
}

// extendSelection adjusts the selection or creates a selection. Negative values make the selection go left and
// positive values make the selection go right.
// "Go" in context of a selection means retracting or extending depending on which side the cursor is on.
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
	field.copy("primary", false)
}

// MoveCursorUp moves the cursor up one line.
//
// If extendSelection is true, the selection is either extended up if the cursor is at the beginning of the selection or
// retracted from the bottom if the cursor is at the end of the selection.
func (field *InputArea) MoveCursorUp(extendSelection bool) {
	pX, pY := field.cursorOffsetX, field.cursorOffsetY
	if field.cursorOffsetY > 0 {
		field.cursorOffsetY--
		lineWidth := iaStringWidth(field.lines[field.cursorOffsetY])
		if lineWidth < field.cursorOffsetX {
			field.cursorOffsetX = lineWidth
		}
	} else {
		field.cursorOffsetX = 0
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
	prevOffsetW := field.cursorOffsetW
	field.recalculateCursorOffset()
	if field.cursorOffsetW == prevOffsetW && field.pressKeyUpAtStart != nil {
		field.pressKeyUpAtStart()
	}
}

// MoveCursorDown moves the cursor down one line.
//
// If extendSelection is true, the selection is either extended down if the cursor is at the end of the selection or
// retracted from the top if the cursor is at the beginning of the selection.
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
	prevOffsetW := field.cursorOffsetW
	field.recalculateCursorOffset()
	if field.cursorOffsetW == prevOffsetW && field.pressKeyDownAtEnd != nil {
		field.pressKeyDownAtEnd()
	}
}

// SetCursorPos sets the X and Y cursor offsets.
func (field *InputArea) SetCursorPos(x, y int) {
	field.cursorOffsetX = x
	field.cursorOffsetY = y
	field.selectionStartW = -1
	field.selectionEndW = -1
	if field.cursorOffsetY > len(field.lines) {
		field.cursorOffsetY = len(field.lines) - 1
	}
	field.recalculateCursorOffset()
}

func (field *InputArea) GetCursorPos() (int, int) {
	return field.cursorOffsetX, field.cursorOffsetY
}

// SetCursorOffset sets the runewidth cursor offset.
func (field *InputArea) SetCursorOffset(offset int) {
	field.cursorOffsetW = offset
	field.selectionStartW = -1
	field.selectionEndW = -1
}

func (field *InputArea) GetCursorOffset() int {
	return field.cursorOffsetW
}

func (field *InputArea) SetSelection(start, end int) {
	field.selectionStartW = start
	field.selectionEndW = end
}

func (field *InputArea) GetSelectedText() string {
	leftLeft := iaSubstringBefore(field.text, field.selectionStartW)
	rightLeft := iaSubstringBefore(field.text, field.selectionEndW)
	return rightLeft[len(leftLeft):]
}

func (field *InputArea) GetSelection() (int, int) {
	return field.selectionStartW, field.selectionEndW
}

func (field *InputArea) ClearSelection() {
	field.selectionStartW = -1
	field.selectionEndW = -1
}

// findWordAt finds the word around the given runewidth offset in the given string.
//
// Returns the start and end index of the word.
func findWordAt(line string, x int) (beforePos, afterPos int) {
	before := iaSubstringBefore(line, x)
	after := line[len(before):]
	afterBound := boundaryPattern.FindStringIndex(after)
	if afterBound != nil {
		afterPos = afterBound[0]
	} else {
		afterPos = len(after)
	}
	afterPos += len(before)
	beforeBounds := boundaryPattern.FindAllStringIndex(before, -1)
	if len(beforeBounds) > 0 {
		beforeBound := beforeBounds[len(beforeBounds)-1]
		beforePos = beforeBound[1]
	} else {
		beforePos = 0
	}
	return
}

// startSelectionStreak selects the current word or line for double and triple clicks (respectively).
func (field *InputArea) startSelectionStreak(x, y int) {
	field.cursorOffsetY = y
	if field.cursorOffsetY > len(field.lines) {
		field.cursorOffsetY = len(field.lines) - 1
	} else if len(field.lines) == 0 {
		return
	}
	line := field.lines[field.cursorOffsetY]
	fullLine := (field.clickStreak-2)%2 == 1
	if fullLine {
		field.cursorOffsetX = iaStringWidth(line)

		field.recalculateCursorOffset()

		field.selectionStartW = field.cursorOffsetW - field.cursorOffsetX
		field.selectionEndW = field.cursorOffsetW
	} else {
		beforePos, afterPos := findWordAt(line, x)
		field.cursorOffsetX = iaStringWidth(line[:afterPos])

		field.recalculateCursorOffset()

		wordWidth := iaStringWidth(line[beforePos:afterPos])
		field.selectionStartW = field.cursorOffsetW - wordWidth
		field.selectionEndW = field.cursorOffsetW

		field.selectionStreakStartWStart = field.selectionStartW
		field.selectionStreakStartWEnd = field.selectionEndW
		field.selectionStreakStartXStart = field.cursorOffsetX - wordWidth
	}

	field.selectionStreakStartY = field.cursorOffsetY
	field.copy("primary", false)
}

// ExtendSelection extends the selection as if the user dragged their mouse to the given coordinates.
func (field *InputArea) ExtendSelection(x, y int) {
	field.cursorOffsetY = y
	if field.cursorOffsetY > len(field.lines) {
		field.cursorOffsetY = len(field.lines) - 1
	}
	if field.clickStreak <= 1 {
		field.cursorOffsetX = x
	} else if (field.clickStreak-2)%2 == 0 {
		if field.lastClickY == y && x >= field.lastWordSelectionExtendXStart && x <= field.lastWordSelectionExtendXEnd {
			return
		}
		line := field.lines[field.cursorOffsetY]
		beforePos, afterPos := findWordAt(line, x)
		field.lastWordSelectionExtendXStart = beforePos
		field.lastWordSelectionExtendXEnd = afterPos
		if y < field.selectionStreakStartY || (y == field.selectionStreakStartY && x < field.selectionStreakStartXStart) {
			field.cursorOffsetW = field.selectionStartW
			field.selectionEndW = field.selectionStreakStartWEnd
			field.cursorOffsetX = iaStringWidth(line[:beforePos])
		} else {
			field.cursorOffsetW = field.selectionEndW
			field.selectionStartW = field.selectionStreakStartWStart
			field.cursorOffsetX = iaStringWidth(line[:afterPos])
		}
	} else {
		if field.lastClickY == y {
			return
		}
		if field.cursorOffsetY == field.selectionStreakStartY {
			// Special case to not mess up stuff when dragging mouse over selection streak start.
			line := field.lines[field.cursorOffsetY]
			field.cursorOffsetX = iaStringWidth(line)
			field.recalculateCursorOffset()
			field.selectionStartW = field.cursorOffsetW - field.cursorOffsetX
			field.selectionEndW = field.cursorOffsetW
			return
		} else if field.cursorOffsetY < field.selectionStreakStartY {
			field.cursorOffsetW = field.selectionStartW
			field.cursorOffsetX = 0
		} else {
			field.cursorOffsetW = field.selectionEndW
			line := field.lines[field.cursorOffsetY]
			field.cursorOffsetX = iaStringWidth(line)
		}
	}
	prevOffset := field.cursorOffsetW
	field.recalculateCursorOffset()
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
	field.copy("primary", false)
}

// RemoveNextCharacter removes the character after the cursor.
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

// RemovePreviousWord removes the word before the cursor.
func (field *InputArea) RemovePreviousWord() {
	left := iaSubstringBefore(field.text, field.cursorOffsetW)
	replacement := lastWord.ReplaceAllString(left, "")
	field.text = replacement + field.text[len(left):]
	field.cursorOffsetW = iaStringWidth(replacement)
}

// RemoveSelection removes the selected content.
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

// RemovePreviousCharacter removes the character before the cursor.
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

// Clear clears the input area.
func (field *InputArea) Clear() {
	field.text = ""
	field.cursorOffsetW = 0
	field.cursorOffsetX = 0
	field.cursorOffsetY = 0
	field.selectionEndW = -1
	field.selectionStartW = -1
	field.viewOffsetY = 0
}

// SelectAll extends the selection to cover all text in the input area.
func (field *InputArea) SelectAll() {
	field.selectionStartW = 0
	field.selectionEndW = iaStringWidth(field.text)
	field.cursorOffsetW = field.selectionEndW
	field.copy("primary", false)
}

// handleInputChanges calls the text change handler and makes sure
// offsets are valid after a change in the text of the input area.
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

// OnPasteEvent handles a terminal bracketed paste event.
func (field *InputArea) OnPasteEvent(event PasteEvent) bool {
	var left, right string
	if field.selectionEndW != -1 {
		left = iaSubstringBefore(field.text, field.selectionStartW)
		rightLeft := iaSubstringBefore(field.text, field.selectionEndW)
		right = field.text[len(rightLeft):]
		field.cursorOffsetW = field.selectionStartW
	} else {
		left = iaSubstringBefore(field.text, field.cursorOffsetW)
		right = field.text[len(left):]
	}
	oldText := field.text
	field.text = left + event.Text() + right
	field.cursorOffsetW += iaStringWidth(event.Text())
	field.handleInputChanges(oldText)
	field.selectionEndW = -1
	field.selectionStartW = -1
	field.snapshot()
	return true
}

// Paste reads the clipboard and inserts the content at the cursor position.
func (field *InputArea) Paste() {
	text, _ := clipboard.ReadAll("clipboard")
	field.OnPasteEvent(tcell.NewEventPaste(text, ""))
}

// Copy copies the currently selected content onto the clipboard.
func (field *InputArea) Copy() {
	field.copy("clipboard", false)
}

func (field *InputArea) Cut() {
	field.copy("clipboard", true)
}

func (field *InputArea) copy(selection string, cut bool) {
	if !field.copySelection && selection == "primary" {
		return
	} else if field.selectionEndW == -1 {
		return
	}
	left := iaSubstringBefore(field.text, field.selectionStartW)
	rightLeft := iaSubstringBefore(field.text, field.selectionEndW)
	text := rightLeft[len(left):]
	_ = clipboard.WriteAll(text, selection)
	if cut {
		field.text = left + field.text[len(rightLeft):]
		field.cursorOffsetW = field.selectionStartW
		field.selectionStartW = -1
		field.selectionEndW = -1
	}
}

// OnKeyEvent handles a terminal key press event.
func (field *InputArea) OnKeyEvent(event KeyEvent) bool {
	hasMod := func(mod tcell.ModMask) bool {
		return event.Modifiers()&mod != 0
	}
	oldText := field.text

	doSnapshot := false
	// Process key event.
	switch event.Key() {
	case tcell.KeyRune:
		field.TypeRune(event.Rune())
		doSnapshot = true
	case tcell.KeyEnter:
		field.TypeRune('\n')
		doSnapshot = true
	case tcell.KeyLeft:
		field.MoveCursorLeft(hasMod(tcell.ModCtrl), hasMod(tcell.ModShift))
	case tcell.KeyRight:
		field.MoveCursorRight(hasMod(tcell.ModCtrl), hasMod(tcell.ModShift))
	case tcell.KeyUp:
		field.MoveCursorUp(hasMod(tcell.ModShift))
	case tcell.KeyDown:
		field.MoveCursorDown(hasMod(tcell.ModShift))
	case tcell.KeyDelete:
		field.RemoveNextCharacter()
		doSnapshot = true
	case tcell.KeyBackspace:
		field.RemovePreviousWord()
		doSnapshot = true
	case tcell.KeyBackspace2:
		field.RemovePreviousCharacter()
		doSnapshot = true
	case tcell.KeyTab:
		if field.tabComplete != nil {
			field.tabComplete(field.text, field.cursorOffsetW)
		}
	default:
		if field.vimBindings {
			switch event.Key() {
			case tcell.KeyCtrlU:
				field.Clear()
				doSnapshot = true
			case tcell.KeyCtrlW:
				field.RemovePreviousWord()
				doSnapshot = true
			default:
				return false
			}
		} else {
			switch event.Key() {
			case tcell.KeyCtrlA:
				field.SelectAll()
			case tcell.KeyCtrlZ:
				field.Undo()
			case tcell.KeyCtrlY:
				field.Redo()
			case tcell.KeyCtrlC:
				field.Copy()
			case tcell.KeyCtrlV:
				field.Paste()
				return true
			case tcell.KeyCtrlX:
				field.Cut()
			default:
				return false
			}
		}
	}
	field.handleInputChanges(oldText)
	if doSnapshot {
		field.snapshot()
	}
	return true
}

// Focus marks the input area as focused.
func (field *InputArea) Focus() {
	field.focused = true
}

// Blur marks the input area as not focused.
func (field *InputArea) Blur() {
	field.focused = false
}

// OnMouseEvent handles a terminal mouse event.
func (field *InputArea) OnMouseEvent(event MouseEvent) bool {
	switch event.Buttons() {
	case tcell.Button1:
		cursorX, cursorY := event.Position()
		cursorY += field.viewOffsetY
		now := millis()
		sameCell := field.lastClickX == cursorX && field.lastClickY == cursorY
		if !event.HasMotion() {
			withinTimeout := now < field.lastClick+field.doubleClickTimeout
			if field.clickStreak > 0 && sameCell && withinTimeout {
				field.clickStreak++
			} else {
				field.clickStreak = 1
			}
			if field.clickStreak <= 1 {
				field.SetCursorPos(cursorX, cursorY)
			} else {
				field.startSelectionStreak(cursorX, cursorY)
			}
			field.lastClick = now
			field.lastClickX = cursorX
			field.lastClickY = cursorY
		} else {
			if sameCell {
				return false
			}
			field.ExtendSelection(cursorX, cursorY)
		}
	case tcell.WheelDown:
		field.viewOffsetY += 3
		field.cursorOffsetY += 3
		field.recalculateCursorOffset()
	case tcell.WheelUp:
		field.viewOffsetY -= 3
		field.cursorOffsetY -= 3
		field.recalculateCursorOffset()
	default:
		return false
	}
	return true
}
