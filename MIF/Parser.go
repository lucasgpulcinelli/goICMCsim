package MIF

import (
	"fmt"
	"io"
)

type Parser struct {
	l          *Lexer
	depth      int64
	width      int64
	dataFormat Format
	addrFormat Format

	dataArray []byte
}

func NewParser(rd io.Reader) *Parser {
	l := NewLexer(rd)
	return &Parser{l: l}
}

func (p *Parser) GetDimensions() (int64, int64) {
	return p.width, p.depth
}

func (p *Parser) GetData() []byte {
	return p.dataArray
}

func (p *Parser) newError(cause string) error {
	l, c := p.l.GetPosition()
	return MIFError{"parser", l, c, cause}
}

func (p *Parser) expect(expected []Token) (Token, error) {
	tok, err := p.l.NextToken()
	if err != nil {
		return TokNone, err
	}

	for _, ex := range expected {
		if tok == ex {
			return tok, nil
		}
	}

	p.l.UnReadToken()
	return TokNone, p.newError(
		fmt.Sprintf("unexpected token %v in input, wanted %v", tok, expected),
	)
}

func (p *Parser) readAddress() (int64, error) {
	if _, err := p.expect([]Token{TokNumber}); err != nil {
		return 0, err
	}
	v, err := p.readWithFormat(p.l.GetData(), p.addrFormat)
	if err != nil {
		return 0, err
	}
	return v, nil
}
