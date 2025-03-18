package sqltest

import "bytes"

type position struct {
	line  int // line number
	index int // byte index
}

func (a position) add(b position) position {
	return position{
		line:  a.line + b.line,
		index: a.index + b.index,
	}
}

func pos(index int) position {
	return position{index: index}
}

type QueryDelimiter func([]byte) position

func defaultQueryDelimiter(src []byte) position {
	pos := pos(len(src)) // by default index is right after the end
	var off int
	for {
		nl := bytes.IndexByte(src, '\n')
		if nl < 0 {
			if i := delimeterBackward(src, len(src)-1); i >= 0 {
				pos.index = off + i
			}
			return pos
		}
		if i := delimeterBackward(src, nl-1); i >= 0 {
			pos.index = off + i
			return pos
		}
		pos.line++
		off += nl + 1
		if pos.index <= off {
			return pos
		}
		src = src[nl+1:]
	}
}

func delimeterBackward(src []byte, i int) int {
	for ; i >= 0; i-- {
		switch src[i] {
		case ' ', '\t':
		case ';':
			return i
		default:
			return -1
		}
	}
	return -1
}

func skipEmptyLines(src []byte) position {
	var pos position
	for i, c := range src {
		switch c {
		case '\n':
			pos.line++
		case ' ', '\t', ';':
		default:
			pos.index = i
			return pos
		}
	}
	pos.index = len(src)
	return pos
}

type comment int

const (
	commentNone comment = iota
	commentLine
	commentBlock
)

type commentChar int

const (
	commentCharNone commentChar = iota
	commentLineStart
	commentBlockStart
	commentBlockEnd
)

func skipCommentaries(src []byte) position {
	var pos position
	var com comment
	var cc commentChar
	for i, c := range src {
		switch cc {
		case commentBlockStart:
			if c == '*' {
				com = commentBlock
			}
			if c != '/' {
				cc = commentCharNone
			}
		case commentBlockEnd:
			if c == '/' {
				com = commentNone
			}
			if c != '*' {
				cc = commentCharNone
			}
		case commentLineStart:
			if c == '-' {
				com = commentLine
			}
			cc = commentCharNone
		case commentCharNone:
			switch c {
			case '/':
				if com == commentNone {
					cc = commentBlockStart
				}
			case '*':
				if com == commentBlock {
					cc = commentBlockEnd
				}
			case '-':
				if com == commentNone {
					cc = commentLineStart
				}
			case '\n':
				pos.line++
				if com == commentLine {
					com = commentNone
				}
			default:
				if com == commentNone {
					pos.index = i
					return pos
				}
			}
		}
	}
	pos.index = len(src)
	return pos
}
