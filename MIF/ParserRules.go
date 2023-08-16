package MIF

import (
	"strconv"
)

// MIF -> header data EOF
func (p *Parser) Parse() error {
	var err error

	if err = p.header(); err != nil {
		return err
	}

	if p.addrFormat == FormatNone {
		return p.newError("ADDRESS_RADIX not defined before CONTENT")
	}
	if p.dataFormat == FormatNone {
		return p.newError("DATA_RADIX not defined before CONTENT")
	}
	if p.depth <= 0 {
		return p.newError("DEPTH undefined before CONTENT or with invalid value")
	}
	if p.width <= 0 {
		return p.newError("WIDTH undefined before CONTENT or with invalid value")
	}

	if err = p.data(); err != nil {
		return err
	}

	if _, err = p.expect([]Token{TokEOF}); err != nil {
		return err
	}

	return nil
}

// header -> declaration header | eps
func (p *Parser) header() error {
	for {
		tok, err := p.l.NextToken()
		if err != nil {
			return err
		}
		p.l.UnReadToken()

		if tok != TokIdent {
			break
		}
		if err = p.declaration(); err != nil {
			return err
		}
	}

	return nil
}

// declaration -> TokIdent TokEq (TokIdent|TokNumber) TokStmtEnd
func (p *Parser) declaration() (err error) {
	if _, err = p.expect([]Token{TokIdent}); err != nil {
		return
	}
	ident := p.l.GetData()
	if _, err = p.expect([]Token{TokEq}); err != nil {
		return
	}
	if _, err = p.expect([]Token{TokIdent, TokNumber}); err != nil {
		return
	}
	value := p.l.GetData()
	if _, err = p.expect([]Token{TokStmtEnd}); err != nil {
		return
	}

	switch ident {
	case "DEPTH":
		p.depth, err = strconv.ParseInt(value, 10, 64)
	case "WIDTH":
		p.width, err = strconv.ParseInt(value, 10, 64)
	case "DATA_RADIX":
		p.dataFormat, err = p.toFormat(value)
	case "ADDRESS_RADIX":
		p.addrFormat, err = p.toFormat(value)
	default:
		err = p.newError("Unexpected identifier")
	}
	return
}

// data -> TokContent TokBegin addressDefs TokEnd TokStmtEnd
func (p *Parser) data() error {
	var err error

	if _, err = p.expect([]Token{TokContent}); err != nil {
		return err
	}
	if _, err = p.expect([]Token{TokBegin}); err != nil {
		return err
	}

	if err = p.addressDefs(); err != nil {
		return err
	}

	if _, err = p.expect([]Token{TokEnd}); err != nil {
		return err
	}
	if _, err = p.expect([]Token{TokStmtEnd}); err != nil {
		return err
	}

	return nil
}

// addressDefs -> definition addressDefs | eps
func (p *Parser) addressDefs() error {
	p.dataArray = make([]byte, p.width*p.depth/8)

	for {
		tok, err := p.l.NextToken()
		if err != nil {
			return err
		}
		p.l.UnReadToken()

		if tok != TokOpen && tok != TokNumber {
			break
		}
		if err = p.definition(); err != nil {
			return err
		}
	}
	return nil
}

// definition -> address TokColon value TokStmtEnd
func (p *Parser) definition() error {
	start, end, err := p.address()
	if err != nil {
		return err
	}
	if _, err = p.expect([]Token{TokColon}); err != nil {
		return err
	}
	values, err := p.value()
	if err != nil {
		return err
	}
	if _, err = p.expect([]Token{TokStmtEnd}); err != nil {
		return err
	}

	k := 0
	for i := start; i <= end; i++ {
		kvalue := values[k]
		for j := int64(0); j < p.width/8; j++ {
			p.dataArray[i*p.width/8+j] = byte(kvalue >> (p.width - (j+1)*8)) 
		}
		k = (k + 1) % len(values)
	}

	return nil
}

// address -> (TokOpen TokNumber TokRange TokNumber TokClose | TokNumber)
func (p *Parser) address() (int64, int64, error) {
	var start, end int64

	tok, err := p.expect([]Token{TokOpen, TokNumber})
	if err != nil {
		return 0, 0, err
	}

	if tok == TokNumber {
		if start, err = p.readWithFormat(p.l.GetData(), p.addrFormat); err != nil {
			return 0, 0, err
		}
		return start, start, nil
	}

	if start, err = p.readAddress(); err != nil {
		return 0, 0, err
	}
	if start < 0 || start > p.depth {
		return 0, 0, p.newError("address must be between 0 and DEPTH")
	}

	if tok, err = p.expect([]Token{TokRange}); err != nil {
		return 0, 0, err
	}

	if end, err = p.readAddress(); err != nil {
		return 0, 0, err
	}
	if end < 0 || end > p.depth {
		return 0, 0, p.newError("address must be between 0 and DEPTH")
	}
	if end <= start {
		return 0, 0, p.newError("end address in range must be greater than start")
	}

	if tok, err = p.expect([]Token{TokClose}); err != nil {
		return 0, 0, err
	}

	return start, end, err
}

// value -> TokNumber opt_itervalue
// opt_itervalue -> TokNumber opt_itervalue | eps
func (p *Parser) value() ([]int64, error) {
	var err error

	if _, err = p.expect([]Token{TokNumber}); err != nil {
		return nil, err
	}

	ret := make([]int64, 1)
	if ret[0], err = p.readWithFormat(p.l.GetData(), p.dataFormat); err != nil {
		return nil, err
	}

	for {
		tok, err := p.l.NextToken()
		if err != nil {
			return nil, err
		}
		if tok != TokNumber {
			p.l.UnReadToken()
			break
		}
		v, err := p.readWithFormat(p.l.GetData(), p.dataFormat)
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}

	return ret, nil
}
