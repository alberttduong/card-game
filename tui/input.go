package tui 

import (
	"fmt"
	"strings"
)

const (
	TextBoxWidth = FieldLength - 2
)

type Input struct {
	leftAlign bool
	Active bool
	text string
}

func (i *Input) Backspace() {
	if l := len(i.text); l != 0 {
		i.text = i.text[:l - 1] 
	}
}

func (i *Input) Space() {
	i.text += " " 
}

func (i *Input) Reset() string {
	defer func() {
		i.text = ""
		i.Active = false
	}()

	return i.text
}

func (i *Input) AddKey(key rune) {
	if key >= 'a' && key <= 'z' ||
	   key >= '0' && key <= '9' {
		i.text += string(key)
	}
}

func (i Input) Textbox() []string { 
	box := func() []string {
		return []string {
			fmt.Sprintf("┌%s┐", strings.Repeat("─", TextBoxWidth)),
			fmt.Sprintf("│%*s│", TextBoxWidth, i.text), 
			fmt.Sprintf("└%s┘", strings.Repeat("─", TextBoxWidth)),
		}
	}

	if i.Active {
		return yellow(box())
	}

	if i.leftAlign {
		b := box()
		b[1] = fmt.Sprintf("│%-*s│", TextBoxWidth, i.text)
		return b
	}
	return box()
}
