package screen

import (
	"encoding/json"
	"strings"
)

type Op uint8

const (
	OpResize Op = iota + 1
	OpClear
	OpKeyboard
	OpCursor
	OpStyle
	OpPut
	OpPutRep
	OpScroll
	OpFlush
	OpEnd
)

type Attr uint8

const (
	AttrReverse Attr = 1 << iota
	AttrItalic
	AttrBold
	AttrUnderline
	AttrUndercurl
	AttrEnd
)

type Mode uint8

const (
	ModeBusy Mode = 1 << iota
	ModeMouseOn
	ModeNormal
	ModeInsert
	ModeReplace
	ModeEnd
)

var JSConst []byte

func init() {
	consts := map[string]int{}

	o := OpEnd
	for o > 0 {
		o--
		s := o.String()
		if strings.ContainsRune(s, '(') {
			continue
		}
		consts[s] = int(o)
	}

	a := AttrEnd
	for a > 0 {
		a--
		s := a.String()
		if strings.ContainsRune(s, '(') {
			continue
		}
		consts[s] = int(a)
	}

	m := ModeEnd
	for m > 0 {
		m--
		s := m.String()
		if strings.ContainsRune(s, '(') {
			continue
		}
		consts[s] = int(m)
	}

	data, err := json.Marshal(consts)
	if err == nil {
		JSConst = append([]byte("var nmux="), data...)
	}
}
