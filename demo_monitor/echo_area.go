package main

import (
	"bytes"
	"unicode/utf8"

	"github.com/gizak/termui"
)

type EchoArea struct {
	termui.Block
	Label string
	data  []string
}

func NewEchoArea(height int) *EchoArea {
	e := &EchoArea{
		Block: *termui.NewBlock(),
		data:  make([]string, 0, height),
	}
	e.Width = 12
	e.Height = height + 2

	return e
}

func (e *EchoArea) AppendLine(line string) {
	if len(e.data) > e.Height-3 {
		e.data = e.data[1:]
	}
	e.data = append(e.data, line)
}

func (e *EchoArea) Write(data []byte) (int, error) {
	line := bytes.NewBufferString("")
	for _, b := range data {
		if b != '\n' {
			line.WriteByte(b)
			continue
		}
		e.AppendLine(line.String())
		line.Reset()
	}
	// Update screen
	termui.Render(e)
	return len(data), nil
}

func (e *EchoArea) Buffer() termui.Buffer {

	buf := e.Block.Buffer()

	bounds := e.InnerBounds()

	for j, line := range e.data {
		for i, w := 0, 0; i < len(line); i += w {
			if i >= e.Width {
				break
			}
			var runeValue rune
			runeValue, w = utf8.DecodeRuneInString(line[i:])
			c := termui.Cell{
				Ch: runeValue,
			}

			buf.Set(bounds.Min.X+i, bounds.Min.Y+j, c)
		}
	}

	// print the label
	for i, w := 0, 0; i < len(e.Label); i += w {
		var runeValue rune
		runeValue, w = utf8.DecodeRuneInString(e.Label[i:])
		c := termui.Cell{
			Ch: runeValue,
			Fg: termui.ColorGreen,
		}

		buf.Set(bounds.Min.X+i, bounds.Min.Y-1, c)
	}

	return buf
}
