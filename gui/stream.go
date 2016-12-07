package gui

import "github.com/tweekmonster/nmux/screen"

type StreamReader struct {
	Data   []byte
	cursor int
}

func (s *StreamReader) Remaining() int {
	return len(s.Data) - s.cursor
}

func (s *StreamReader) ReadOp() screen.Op {
	return screen.Op(s.ReadUint8())
}

func (s *StreamReader) ReadUint8() uint8 {
	b := s.Data[s.cursor]
	s.cursor++
	return b
}

func (s *StreamReader) ReadInt16() int16 {
	return int16(s.ReadUint8())<<8 | int16(s.ReadUint8())
}

func (s *StreamReader) ReadUint24() int {
	return int(s.ReadUint8())<<16 | int(s.ReadUint8())<<8 | int(s.ReadUint8())
}

func (s *StreamReader) ReadEint32() int {
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
	length := s.ReadEint32()
	out := ""
	for i := 0; i < length; i++ {
		out += string(s.ReadEint32())
	}

	return out
}
