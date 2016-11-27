package gui

import "github.com/tweekmonster/nmux/screen"

type StreamReader struct {
	Data   []byte
	cursor int
}

func (s *StreamReader) ReadOp() screen.Op {
	op := screen.Op(s.Data[s.cursor])
	s.cursor++
	return op
}

func (s *StreamReader) Eint32() int {
	i := int(s.Data[s.cursor])

	if (i & 0x80) == 0 {
		s.cursor++
		return i
	}

	i = (i & 0x7f) | int(s.Data[s.cursor+1])<<7
	if (i & 0x4000) == 0 {
		s.cursor += 2
		return i
	}

	i = (i & 0x3fff) | int(s.Data[s.cursor+2])<<14
	if (i & 0x200000) == 0 {
		s.cursor += 3
		return i
	}

	i = (i & 0x1fffff) | int(s.Data[s.cursor+3])<<21
	if (i & 0x10000000) == 0 {
		s.cursor += 4
		return i
	}

	i = (i & 0xfffffff) | int(s.Data[s.cursor+4])<<28
	s.cursor += 5
	return i
}

func (s *StreamReader) ReadString() string {
	length := s.Eint32()
	out := ""
	for i := 0; i < length; i++ {
		out += string(s.Eint32())
	}

	return out
}
