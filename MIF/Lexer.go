package MIF

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

type Lexer struct {
	stream   *bufio.Reader
	line     int
	col      int
	lastTok  Token
	rewinded bool

	dataValue string
}

func NewLexer(rd io.Reader) *Lexer {
	bufrd := bufio.NewReader(rd)
	return &Lexer{stream: bufrd, line: 1, col: 1}
}

func (l *Lexer) GetData() string {
	return l.dataValue
}

func (l *Lexer) GetPosition() (int, int){
  return l.line, l.col
}

func (l *Lexer) newError(cause string) error {
	return MIFError{"lexer", l.line, l.col, cause}
}

func (l *Lexer) readByte() (byte, error) {
	c, err := l.stream.ReadByte()
	if err != nil {
		return byte('\x00'), err
	}
	if c == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return c, nil
}

func (l *Lexer) readUntil(end byte) (err error) {
	for c := byte(' '); c != end; c, err = l.readByte() {
		if err != nil {
			return
		}
	}
	return
}

func (l *Lexer) readIdent() (err error) {
  var c byte

	l.dataValue = ""
	for {
		if c, err = l.readByte(); err != nil {
			break
		}
		if !unicode.IsLetter(rune(c)) && c != '_' && !unicode.IsNumber(rune(c)) {
			err = l.stream.UnreadByte()
			break
		}

		l.dataValue += string(c)
	}

	return err
}

func (l *Lexer) nextToken() (Token, error) {
	var err error

	c := byte(' ')
	for unicode.IsSpace(rune(c)) {
		c, err = l.readByte()
		if err != nil {
			return TokNone, err
		}
	}

	if unicode.IsLetter(rune(c)) || unicode.IsNumber(rune(c)) || c == '_' {
		if err = l.stream.UnreadByte(); err != nil {
			return TokNone, err
		}
		if err = l.readIdent(); err != nil {
			return TokNone, err
		}

		if l.dataValue == "CONTENT" {
			return TokContent, nil
		}
		if l.dataValue == "BEGIN" {
			return TokBegin, nil
		}
		if l.dataValue == "END" {
			return TokEnd, nil
		}

		if unicode.IsNumber(rune(c)) {
			return TokNumber, nil
		} else {
			return TokIdent, nil
		}
	}

	switch c {
	case '%':
		if err = l.readUntil('%'); err != nil {
			return TokNone, err
		}
		return l.nextToken()
	case '-':
		if c, err = l.readByte(); err != nil {
			return TokNone, err
		}
		if c != '-' {
			return TokNone, l.newError("expected '-'")
		}
		if err = l.readUntil('\n'); err != nil {
			return TokEnd, err
		}

		return l.nextToken()
	case ';':
		return TokStmtEnd, nil
	case ':':
		return TokColon, nil
	case ']':
		return TokClose, nil
	case '[':
		return TokOpen, nil
	case '=':
		return TokEq, nil
	case '.':
		if c, err = l.readByte(); err != nil {
			return TokNone, err
		}
		if c != '.' {
			return TokNone, l.newError("expected '.'")
		}
		return TokRange, nil
	}

	return TokNone, l.newError(fmt.Sprintf("invalid character %c in input", c))
}

func (l *Lexer) NextToken() (Token, error) {
	if l.rewinded {
		l.rewinded = false
		return l.lastTok, nil
	}

	tok, err := l.nextToken()
  if err == io.EOF {
    tok = TokEOF
    err = nil
  }

	l.lastTok = tok

	return tok, err
}

func (l *Lexer) UnReadToken() {
	l.rewinded = true
}
