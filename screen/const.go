package screen

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Op is the screen operation type.
type Op uint8

// Screen operations.
const (
	OpResize Op = iota + 1
	OpClear
	OpKeyboard
	OpCursor
	OpPalette
	OpStyle
	OpPut
	OpPutRep
	OpTitle
	OpIcon
	OpBell
	OpScroll
	OpFlush
	OpLog
	OpEnd
)

// Attr is a set of screen cell attributes.
type Attr uint8

// Cell attributes
const (
	AttrReverse Attr = 1 << iota
	AttrItalic
	AttrBold
	AttrUnderline
	AttrUndercurl
	AttrEnd
)

// Mode is the Neovim mode and other state information.
type Mode uint8

// Modes
const (
	ModeBusy Mode = 1 << iota
	ModeMouseOn
	ModeNormal
	ModeInsert
	ModeReplace
	ModeRedraw
	ModeEnd
)

// JSConst will contain a javascript representation of the constants above.
var JSConst []byte

func init() {
	consts := map[string]interface{}{}

	o := OpEnd
	for o > 0 {
		o--
		s := fmt.Sprintf("%s", o)
		if strings.ContainsRune(s, '(') {
			continue
		}

		consts[s] = o
		consts[fmt.Sprintf("o%d", o)] = s
	}

	a := AttrEnd
	for a > 0 {
		a--
		s := fmt.Sprintf("%s", a)
		if strings.ContainsRune(s, '(') {
			continue
		}
		consts[s] = a
		consts[fmt.Sprintf("a%d", a)] = s
	}

	m := ModeEnd
	for m > 0 {
		m--
		s := fmt.Sprintf("%s", m)
		if strings.ContainsRune(s, '(') {
			continue
		}
		consts[s] = m
		consts[fmt.Sprintf("m%d", m)] = s
	}

	data, err := json.Marshal(consts)
	if err == nil {
		JSConst = append([]byte("var nmux="), data...)
	}
}
