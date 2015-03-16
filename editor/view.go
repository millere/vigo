// Copyright 2015 Ethan Miller. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package editor

import (
	"fmt"
	"strings"

	"github.com/millere/jk/keys"
	"github.com/nsf/termbox-go"
)

type View struct {
	x, y      int    // (x, y) position of the top left corner of the view
	w, h      int    // width and height of the view
	FirstLine int    // index of the first line
	back      Buffer // the backing buffer being displayed
	C         Cursor
	mode      *Mode
}

type Cursor struct {
	Line, Column int
	color        termbox.Attribute
}

func (e *Editor) ViewWithBuffer(a Buffer, m string, x, y, w, h int) (View, error) {
	mode, ok := e.modes[m]
	if !ok {
		return View{}, fmt.Errorf("Mode \"%v\" does not exist", m)
	}
	return View{
		x:         x,
		y:         y,
		w:         w,
		h:         h,
		back:      a,
		FirstLine: 1,
		C: Cursor{
			Line:   1,
			Column: 1,
			color:  termbox.ColorRed,
		},
		mode: mode,
	}, nil
}

func (a *View) Draw() {
	ClearBox(a.x, a.y, a.w, a.h)
	currentLine, err := a.back.GetLine(a.FirstLine)
	if err != nil {
		// TODO: handle error better
		panic(err)
	}
	for yi := 0; yi < a.h; yi++ {
		offset := 0
		for xi, c := range string(currentLine.Contents) {
			if xi >= a.w {
				break
			}
			if c != '\t' {
				termbox.SetCell(a.x+xi+offset, a.y+yi, c, termbox.ColorDefault, termbox.ColorDefault)
			} else {
				offset += 4
			}

		}
		currentLine = currentLine.next
		if currentLine == nil {
			break
		}
	}
	cursorLine, err := a.back.GetLine(a.C.Line)
	if err != nil {
		panic(err)
	}
	tabsAtCursor := strings.Count(string(cursorLine.Contents[:a.C.Column-1]), "\t")
	termbox.SetCursor(a.x+a.C.Column-1+4*tabsAtCursor, a.y+a.C.Line-1) // context required for humor.
}

// sets the cursor to absolute coordinates in the file
func (a *View) SetCursor(column, row int) {
	target, err := a.back.GetLine(row)
	_ = target
	if err != nil {
		return
	}

	if inBounds(0, 0, len(target.Contents), a.back.Lines(), column, row) {
		a.C.Line = row
		a.C.Column = column
	} else if len(target.Contents) >= len(target.Contents) {
		a.C.Line = row
		a.C.Column = len(target.Contents)
	} else {
		LogItAll.Printf("Position (%d, %d) out of bounds (%d, %d, %d, %d)\n",
			column-1, row-1,
			0, 0,
			len(target.Contents), a.back.Lines(),
		)
	}
}

// moves the cursor relative to where it is now
func (a *View) MoveCursor(dc, dr int) {
	a.SetCursor(a.C.Column+dc, a.C.Line+dr)
}

func inBounds(x, y, w, h, ax, ay int) bool {
	return ax >= x && ax < w && ay >= y && ay < h
}

func (a *View) SetMode(m *Mode) {
	if a.mode.OnExit != nil {
		a.mode.OnExit(a)
	}
	a.mode = m
	if a.mode.OnEnter != nil {
		a.mode.OnEnter(a)
	}
}

func (a *View) Do(k keys.Keypress) error {
	f, ok := a.mode.EventMap[k]
	if ok {
		return f(a, 1)
	} else {
		return nil
	}
}
