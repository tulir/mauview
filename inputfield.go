// mauview - A Go TUI library based on tcell.
// Copyright © 2019 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Based on https://github.com/rivo/tview/blob/master/inputfield.go

package mauview

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"

	"maunium.net/go/tcell"
)

// InputField is a single-line user-editable text field.
//
// Use SetMaskCharacter() to hide input from onlookers (e.g. for password
// input).
type InputField struct {
	// Cursor position
	cursorOffset int
	// Number of characters (from left) to offset rendering.
	viewOffset int

	// The text that was entered.
	text string

	// The text to be displayed in the input area when it is empty.
	placeholder string

	// The background color of the input area.
	fieldBackgroundColor tcell.Color
	// The text color of the input area.
	fieldTextColor tcell.Color
	// The text color of the placeholder.
	placeholderTextColor tcell.Color

	// A character to mask entered text (useful for password fields). A value of 0
	// disables masking.
	maskCharacter rune

	// Whether or not to enable vim-style keybindings.
	vimBindings bool

	// Whether or not the input field is focused.
	focused bool

	// An optional function which is called when the input has changed.
	changed func(text string)

	// An optional function which is called when the user presses tab.
	tabComplete func(text string, pos int)
}

// NewInputField returns a new input field.
func NewInputField() *InputField {
	return &InputField{
		fieldBackgroundColor: Styles.ContrastBackgroundColor,
		fieldTextColor:       Styles.PrimaryTextColor,
		placeholderTextColor: Styles.ContrastSecondaryTextColor,
	}
}

// SetText sets the current text of the input field.
func (field *InputField) SetText(text string) *InputField {
	field.text = text
	if field.changed != nil {
		field.changed(text)
	}
	return field
}

// SetTextAndMoveCursor sets the current text of the input field and moves the cursor with the width difference.
func (field *InputField) SetTextAndMoveCursor(text string) *InputField {
	oldWidth := StringWidth(field.text)
	field.text = text
	newWidth := StringWidth(field.text)
	if oldWidth != newWidth {
		field.cursorOffset += newWidth - oldWidth
	}
	if field.changed != nil {
		field.changed(field.text)
	}
	return field
}

// GetText returns the current text of the input field.
func (field *InputField) GetText() string {
	return field.text
}

// SetPlaceholder sets the text to be displayed when the input text is empty.
func (field *InputField) SetPlaceholder(text string) *InputField {
	field.placeholder = text
	return field
}

// SetBackgroundColor sets the background color of the input area.
func (field *InputField) SetBackgroundColor(color tcell.Color) *InputField {
	field.fieldBackgroundColor = color
	return field
}

// SetTextColor sets the text color of the input area.
func (field *InputField) SetTextColor(color tcell.Color) *InputField {
	field.fieldTextColor = color
	return field
}

// SetPlaceholderTextColor sets the text color of placeholder text.
func (field *InputField) SetPlaceholderTextColor(color tcell.Color) *InputField {
	field.placeholderTextColor = color
	return field
}

// SetMaskCharacter sets a character that masks user input on a screen. A value
// of 0 disables masking.
func (field *InputField) SetMaskCharacter(mask rune) *InputField {
	field.maskCharacter = mask
	return field
}

// SetChangedFunc sets a handler which is called whenever the text of the input
// field has changed. It receives the current text (after the change).
func (field *InputField) SetChangedFunc(handler func(text string)) *InputField {
	field.changed = handler
	return field
}

func (field *InputField) SetTabCompleteFunc(handler func(text string, cursorOffset int)) *InputField {
	field.tabComplete = handler
	return field
}

// prepareText prepares the text to be displayed and recalculates the view and cursor offsets.
func (field *InputField) prepareText(screen Screen) (text string, placeholder bool) {
	width, _ := screen.Size()
	text = field.text
	if len(text) == 0 && len(field.placeholder) > 0 {
		text = field.placeholder
		placeholder = true
	}

	if !placeholder && field.maskCharacter > 0 {
		text = strings.Repeat(string(field.maskCharacter), utf8.RuneCountInString(text))
	}
	textWidth := StringWidth(text)
	if field.cursorOffset >= textWidth {
		width--
	}

	if field.cursorOffset < field.viewOffset {
		field.viewOffset = field.cursorOffset
	} else if field.cursorOffset > field.viewOffset+width {
		field.viewOffset = field.cursorOffset - width
	} else if textWidth-field.viewOffset < width {
		field.viewOffset = textWidth - width
	}

	if field.viewOffset < 0 {
		field.viewOffset = 0
	}

	return
}

// drawText draws the text and the cursor.
func (field *InputField) drawText(screen Screen, text string, placeholder bool) {
	width, _ := screen.Size()
	runes := []rune(text)
	x := 0
	style := tcell.StyleDefault.Foreground(field.fieldTextColor).Background(field.fieldBackgroundColor)
	if placeholder {
		style = style.Foreground(field.placeholderTextColor)
	}
	for pos := field.viewOffset; pos <= width+field.viewOffset && pos < len(runes); pos++ {
		ch := runes[pos]
		w := runewidth.RuneWidth(ch)
		for w > 0 {
			screen.SetContent(x, 0, ch, nil, style)
			x++
			w--
		}
	}
	for ; x <= width; x++ {
		screen.SetContent(x, 0, ' ', nil, style)
	}
}

// Draw draws this primitive onto the screen.
func (field *InputField) Draw(screen Screen) {
	width, height := screen.Size()
	if height < 1 || width < 1 {
		return
	}

	text, placeholder := field.prepareText(screen)
	field.drawText(screen, text, placeholder)
	if field.focused {
		field.setCursor(screen)
	}
}

func (field *InputField) GetCursorOffset() int {
	return field.cursorOffset
}

func (field *InputField) SetCursorOffset(offset int) *InputField {
	if offset < 0 {
		offset = 0
	} else {
		width := StringWidth(field.text)
		if offset >= width {
			offset = width
		}
	}
	field.cursorOffset = offset
	return field
}

// setCursor sets the cursor position.
func (field *InputField) setCursor(screen Screen) {
	width, _ := screen.Size()
	x := field.cursorOffset - field.viewOffset
	if x >= width {
		x = width - 1
	} else if x < 0 {
		x = 0
	}
	screen.ShowCursor(x, 0)
}

var (
	lastWord  = regexp.MustCompile(`\S+\s*$`)
	firstWord = regexp.MustCompile(`^\s*\S+`)
)

func SubstringBefore(s string, w int) string {
	return runewidth.Truncate(s, w, "")
}

func (field *InputField) TypeRune(ch rune) {
	leftPart := SubstringBefore(field.text, field.cursorOffset)
	field.text = leftPart + string(ch) + field.text[len(leftPart):]
	field.cursorOffset += runewidth.RuneWidth(ch)
}

func (field *InputField) MoveCursorLeft(moveWord bool) {
	before := SubstringBefore(field.text, field.cursorOffset)
	if moveWord {
		found := lastWord.FindString(before)
		field.cursorOffset -= StringWidth(found)
	} else if len(before) > 0 {
		beforeRunes := []rune(before)
		char := beforeRunes[len(beforeRunes)-1]
		field.cursorOffset -= runewidth.RuneWidth(char)
	}
}

func (field *InputField) MoveCursorRight(moveWord bool) {
	before := SubstringBefore(field.text, field.cursorOffset)
	after := field.text[len(before):]
	if moveWord {
		found := firstWord.FindString(after)
		field.cursorOffset += StringWidth(found)
	} else if len(after) > 0 {
		char := []rune(after)[0]
		field.cursorOffset += runewidth.RuneWidth(char)
	}
}

func (field *InputField) RemoveNextCharacter() {
	if field.cursorOffset >= StringWidth(field.text) {
		return
	}
	leftPart := SubstringBefore(field.text, field.cursorOffset)
	// Take everything after the left part minus the first character.
	rightPart := string([]rune(field.text[len(leftPart):])[1:])

	field.text = leftPart + rightPart
}

func (field *InputField) Clear() {
	field.text = ""
	field.cursorOffset = 0
	field.viewOffset = 0
}

func (field *InputField) RemovePreviousWord() {
	leftPart := SubstringBefore(field.text, field.cursorOffset)
	rightPart := field.text[len(leftPart):]
	replacement := lastWord.ReplaceAllString(leftPart, "")
	field.text = replacement + rightPart

	field.cursorOffset -= StringWidth(leftPart) - StringWidth(replacement)
}

func (field *InputField) RemovePreviousCharacter() {
	if field.cursorOffset == 0 {
		return
	}
	leftPart := SubstringBefore(field.text, field.cursorOffset)
	rightPart := field.text[len(leftPart):]

	// Take everything before the right part minus the last character.
	leftPartRunes := []rune(leftPart)
	leftPartRunes = leftPartRunes[0 : len(leftPartRunes)-1]
	leftPart = string(leftPartRunes)

	// Figure out what character was removed to correctly decrease cursorOffset.
	removedChar := field.text[len(leftPart) : len(field.text)-len(rightPart)]

	field.text = leftPart + rightPart

	field.cursorOffset -= StringWidth(removedChar)
}

func (field *InputField) handleInputChanges(originalText string) {
	// Trigger changed events.
	if field.text != originalText && field.changed != nil {
		field.changed(field.text)
	}

	// Make sure cursor offset is valid
	if field.cursorOffset < 0 {
		field.cursorOffset = 0
	}
	width := StringWidth(field.text)
	if field.cursorOffset > width {
		field.cursorOffset = width
	}
}

func (field *InputField) OnPasteEvent(event PasteEvent) bool {
	defer field.handleInputChanges(field.text)
	leftPart := SubstringBefore(field.text, field.cursorOffset)
	field.text = leftPart + event.Text() + field.text[len(leftPart):]
	field.cursorOffset += StringWidth(event.Text())
	return true
}

func (field *InputField) Submit(event KeyEvent) bool {
	return true
}

// Global options to specify which of the two backspace key codes should remove the whole previous word.
// If false, only the previous character will be removed with that key code.
var (
	Backspace1RemovesWord = true
	Backspace2RemovesWord = false
)

func (field *InputField) OnKeyEvent(event KeyEvent) bool {
	defer field.handleInputChanges(field.text)

	// Process key event.
	switch key := event.Key(); key {
	case tcell.KeyRune:
		field.TypeRune(event.Rune())
	case tcell.KeyLeft:
		field.MoveCursorLeft(event.Modifiers() == tcell.ModCtrl)
	case tcell.KeyRight:
		field.MoveCursorRight(event.Modifiers() == tcell.ModCtrl)
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
	case tcell.KeyBackspace:
		if Backspace1RemovesWord {
			field.RemovePreviousWord()
		} else {
			field.RemovePreviousCharacter()
		}
	case tcell.KeyBackspace2:
		if Backspace2RemovesWord {
			field.RemovePreviousWord()
		} else {
			field.RemovePreviousCharacter()
		}
	case tcell.KeyTab:
		if field.tabComplete != nil {
			field.tabComplete(field.text, field.cursorOffset)
			return true
		}
		return false
	default:
		return false
	}
	return true
}

func (field *InputField) Focus() {
	field.focused = true
}

func (field *InputField) Blur() {
	field.focused = false
}

func (field *InputField) OnMouseEvent(event MouseEvent) bool {
	if event.Buttons() == tcell.Button1 {
		x, _ := event.Position()
		field.SetCursorOffset(field.viewOffset + x)
		return true
	}
	return false
}
