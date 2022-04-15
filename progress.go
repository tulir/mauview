// mauview - A Go TUI library based on tcell.
// Copyright © 2020 Tulir Asokan
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package mauview

import (
	"math"
	"sync/atomic"
	"time"

	"go.mau.fi/tcell"
)

type ProgressBar struct {
	*SimpleEventHandler
	style    tcell.Style
	progress int32
	max      int

	indeterminate      bool
	indeterminateStart time.Time
}

var _ Component = &ProgressBar{}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		SimpleEventHandler: &SimpleEventHandler{},

		style:         tcell.StyleDefault,
		progress:      0,
		max:           100,
		indeterminate: true,
	}
}

var Blocks = [9]rune{' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉', '█'}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (pb *ProgressBar) SetProgress(progress int) *ProgressBar {
	pb.progress = int32(min(progress, pb.max))
	return pb
}

func (pb *ProgressBar) Increment(increment int) *ProgressBar {
	atomic.AddInt32(&pb.progress, int32(increment))
	return pb
}

func (pb *ProgressBar) SetIndeterminate(indeterminate bool) *ProgressBar {
	pb.indeterminate = indeterminate
	pb.indeterminateStart = time.Now()
	return pb
}

func (pb *ProgressBar) SetMax(max int) *ProgressBar {
	pb.max = max
	pb.progress = int32(min(pb.max, int(pb.progress)))
	return pb
}

// Draw draws this primitive onto the screen.
func (pb *ProgressBar) Draw(screen Screen) {
	width, _ := screen.Size()
	if pb.indeterminate {
		barWidth := width / 6
		pos := int(time.Now().Sub(pb.indeterminateStart).Milliseconds()/200) % (width + barWidth)
		for x := pos - barWidth; x < pos; x++ {
			screen.SetCell(x, 0, pb.style, Blocks[8])
		}
	} else {
		progress := math.Min(float64(pb.progress), float64(pb.max))
		floatingBlocks := progress * (float64(width) / float64(pb.max))
		parts := int(math.Floor(math.Mod(floatingBlocks, 1) * 8))
		blocks := int(math.Floor(floatingBlocks))
		for x := 0; x < blocks; x++ {
			screen.SetCell(x, 0, pb.style, Blocks[8])
		}
		screen.SetCell(blocks, 0, pb.style, Blocks[parts])
	}
}
